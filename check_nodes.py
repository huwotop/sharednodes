#!/usr/bin/env python3
import yaml
import requests
import socket
import time
import base64
import json
from concurrent.futures import ThreadPoolExecutor, as_completed
import argparse
import os
import subprocess
import tempfile
import atexit
import sys
from urllib.parse import quote

MIHOMO_BIN = None
MIHOMO_PROCESS = None
MIHOMO_API_PORT = 9090

TEST_URLS = [
    'http://www.gstatic.com/generate_204',
    'http://cp.cloudflare.com/generate_204',
    'http://connectivitycheck.gstatic.com/generate_204',
    'http://www.google.com/generate_204'
]

def find_mihomo():
    bin_names = ['mihomo', 'clash-meta', 'mihomo-windows-amd64.exe', 'mihomo-windows-386.exe']
    for name in bin_names:
        if os.path.exists(name):
            return os.path.abspath(name)
        for path in os.environ.get('PATH', '').split(os.pathsep):
            bin_path = os.path.join(path, name)
            if os.path.exists(bin_path):
                return bin_path
    return None

def start_mihomo(nodes):
    global MIHOMO_PROCESS, MIHOMO_BIN
    
    if MIHOMO_BIN is None:
        MIHOMO_BIN = find_mihomo()
    
    if MIHOMO_BIN is None:
        print("Error: mihomo binary not found!")
        print("Please download mihomo from https://github.com/MetaCubeX/mihomo/releases")
        sys.exit(1)
    
    config = {
        'mixed-port': 7890,
        'external-controller': f'127.0.0.1:{MIHOMO_API_PORT}',
        'allow-lan': False,
        'mode': 'rule',
        'log-level': 'silent',
        'proxies': nodes,
        'rules': ['MATCH,DIRECT']
    }
    
    fd, config_path = tempfile.mkstemp(suffix='.yaml')
    os.close(fd)
    
    with open(config_path, 'w', encoding='utf-8') as f:
        yaml.dump(config, f)
    
    try:
        MIHOMO_PROCESS = subprocess.Popen(
            [MIHOMO_BIN, '-d', os.path.dirname(config_path), '-f', config_path],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL
        )
        
        time.sleep(3)
        
        for _ in range(15):
            try:
                res = requests.get(f'http://127.0.0.1:{MIHOMO_API_PORT}/proxies', timeout=1)
                if res.status_code == 200:
                    def cleanup():
                        if MIHOMO_PROCESS:
                            MIHOMO_PROCESS.terminate()
                            try:
                                MIHOMO_PROCESS.wait(timeout=2)
                            except:
                                MIHOMO_PROCESS.kill()
                        try:
                            os.unlink(config_path)
                        except:
                            pass
                    atexit.register(cleanup)
                    return True
            except:
                time.sleep(0.5)
        
        MIHOMO_PROCESS.terminate()
        try:
            os.unlink(config_path)
        except:
            pass
        print("Error: mihomo failed to start!")
        sys.exit(1)
        
    except Exception as e:
        print(f"Error: mihomo start error: {e}")
        try:
            os.unlink(config_path)
        except:
            pass
        sys.exit(1)

def test_tcp_connect(server, port, timeout=3):
    try:
        start_time = time.time()
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(timeout)
        result = sock.connect_ex((server, port))
        sock.close()
        if result == 0:
            return True, int((time.time() - start_time) * 1000)
        else:
            return False, 'TCP connection failed'
    except Exception as e:
        return False, str(e)

def test_node(node, timeout=10):
    try:
        node_type = node.get('type', '')
        
        if node_type not in ['vmess', 'vless', 'ss', 'trojan', 'tuic', 'hysteria', 'hysteria2']:
            return None, False, 'Unsupported type'
        
        node_name = node.get('name', 'Unknown')
        server = node.get('server', '')
        port = node.get('port', 0)
        
        tcp_ok, tcp_result = test_tcp_connect(server, port, 3)
        if not tcp_ok:
            return node, False, f'TCP failed: {tcp_result}'
        
        try:
            encoded_name = quote(node_name, safe='')
            for test_url in TEST_URLS:
                try:
                    response = requests.get(
                        f'http://127.0.0.1:{MIHOMO_API_PORT}/proxies/{encoded_name}/delay',
                        params={'timeout': str(timeout * 1000), 'url': test_url},
                        timeout=timeout + 3
                    )
                    
                    if response.status_code == 200:
                        data = response.json()
                        delay = data.get('delay', 0)
                        if delay > 0:
                            return node, True, delay
                except requests.exceptions.Timeout:
                    continue
                except Exception:
                    continue
            
            return node, False, 'All test URLs failed'
                
        except Exception as e:
            return node, False, str(e)
            
    except Exception as e:
        return node, False, str(e)

def node_to_link(node):
    try:
        node_type = node.get('type', '')
        
        if node_type == 'vmess':
            vmess_dict = {
                'v': '2',
                'ps': node.get('name', ''),
                'add': node.get('server', ''),
                'port': str(node.get('port', '')),
                'id': node.get('uuid', ''),
                'aid': str(node.get('alterId', 0)),
                'net': node.get('network', 'tcp'),
                'type': node.get('network', 'none'),
                'host': node.get('ws-headers', {}).get('Host', node.get('ws-path', '')),
                'path': node.get('ws-path', ''),
                'tls': node.get('tls', False) and 'tls' or '',
                'sni': node.get('servername', '')
            }
            vmess_json = json.dumps(vmess_dict, ensure_ascii=False)
            return 'vmess://' + base64.b64encode(vmess_json.encode('utf-8')).decode('utf-8')
        
        elif node_type == 'ss':
            cipher = node.get('cipher', 'aes-256-gcm')
            password = node.get('password', '')
            server = node.get('server', '')
            port = node.get('port', '')
            name = node.get('name', '')
            ss_str = f"{cipher}:{password}@{server}:{port}"
            ss_b64 = base64.urlsafe_b64encode(ss_str.encode('utf-8')).decode('utf-8')
            return f"ss://{ss_b64}#{name}"
        
        elif node_type == 'trojan':
            server = node.get('server', '')
            port = node.get('port', '')
            password = node.get('password', '')
            name = node.get('name', '')
            query = []
            if node.get('sni'):
                query.append(f"sni={node.get('sni')}")
            if node.get('alpn'):
                query.append(f"alpn={','.join(node.get('alpn'))}")
            query_str = '&'.join(query)
            if query_str:
                return f"trojan://{password}@{server}:{port}?{query_str}#{name}"
            else:
                return f"trojan://{password}@{server}:{port}#{name}"
        
        else:
            return None
    except Exception as e:
        print(f"Error converting node to link: {e}")
        return None

def load_subscription(url):
    try:
        headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
        }
        response = requests.get(url, headers=headers, timeout=60)
        response.raise_for_status()
        return yaml.safe_load(response.text)
    except Exception as e:
        print(f"Error loading subscription: {e}")
        return None

def main():
    parser = argparse.ArgumentParser(description='Check Clash subscription nodes and save valid ones (requires mihomo)')
    parser.add_argument('--url', type=str, required=True, help='Clash subscription URL')
    parser.add_argument('--output', type=str, default='valid_nodes.txt', help='Output file')
    parser.add_argument('--timeout', type=int, default=10, help='Connection timeout in seconds')
    parser.add_argument('--threads', type=int, default=20, help='Number of threads')
    
    args = parser.parse_args()
    
    print(f"Loading subscription from: {args.url}")
    config = load_subscription(args.url)
    
    if not config or 'proxies' not in config:
        print("Failed to load subscription or no proxies found")
        return
    
    nodes = config['proxies']
    print(f"Total nodes: {len(nodes)}")
    
    print("Starting mihomo...")
    start_mihomo(nodes)
    print("Testing nodes using mihomo API + TCP pre-check...")
    
    valid_nodes = []
    valid_links = []
    
    with ThreadPoolExecutor(max_workers=args.threads) as executor:
        futures = {executor.submit(test_node, node, args.timeout): node for node in nodes}
        
        for future in as_completed(futures):
            node, is_valid, result = future.result()
            if node is None:
                continue
            if is_valid:
                valid_nodes.append((node, result))
                link = node_to_link(node)
                if link:
                    valid_links.append(link)
                print(f"✓ {node.get('name', 'Unknown')} - {result}ms")
            else:
                print(f"✗ {node.get('name', 'Unknown')} - {result}")
    
    print(f"\nValid nodes: {len(valid_nodes)} / {len(nodes)}")
    
    if valid_links:
        with open(args.output, 'w', encoding='utf-8') as f:
            f.write('\n'.join(valid_links))
        print(f"Valid links saved to: {args.output}")
        
        content = '\n'.join(valid_links)
        b64_content = base64.b64encode(content.encode('utf-8')).decode('utf-8')
        with open(f'{args.output}.b64', 'w', encoding='utf-8') as f:
            f.write(b64_content)
        print(f"Base64 encoded file saved to: {args.output}.b64")

if __name__ == '__main__':
    main()
