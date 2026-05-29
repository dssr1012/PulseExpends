#!/usr/bin/env python3
"""
Check Huawei Cloud ECS instance status in Santiago region
"""

import os
import sys
import json
import subprocess

def check_ecs_via_cli():
    """Check ECS instances using Huawei Cloud CLI if available"""
    try:
        # Try to run huaweicloud CLI
        result = subprocess.run(
            ["huaweicloud", "ecs", "server", "list", "--region", "la-south-2", "--format", "json"],
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode == 0:
            instances = json.loads(result.stdout)
            return instances
        else:
            print(f"❌ Huawei Cloud CLI error: {result.stderr}")
            return None
            
    except FileNotFoundError:
        print("❌ Huawei Cloud CLI not installed")
        return None
    except subprocess.TimeoutExpired:
        print("❌ Huawei Cloud CLI timeout")
        return None
    except json.JSONDecodeError:
        print("❌ Failed to parse Huawei Cloud CLI output")
        return None

def check_ecs_via_ssh(ip_address):
    """Check if ECS instance is reachable via SSH"""
    try:
        # Try to SSH with a simple command
        result = subprocess.run(
            ["ssh", "-o", "ConnectTimeout=5", "-o", "StrictHostKeyChecking=no", 
             f"root@{ip_address}", "echo 'SSH connection successful'"],
            capture_output=True,
            text=True,
            timeout=10
        )
        
        if result.returncode == 0:
            return True
        else:
            return False
            
    except subprocess.TimeoutExpired:
        return False
    except Exception as e:
        print(f"❌ SSH check error: {e}")
        return False

def generate_ssh_key():
    """Generate SSH key pair for ECS access"""
    key_file = "pulse-expends-key.pem"
    
    if os.path.exists(key_file):
        print(f"✅ SSH key already exists: {key_file}")
        
        # Get public key
        pub_key_file = f"{key_file}.pub"
        if os.path.exists(pub_key_file):
            with open(pub_key_file, 'r') as f:
                public_key = f.read().strip()
            print(f"📋 Public key for Huawei Cloud import:")
            print("-" * 60)
            print(public_key)
            print("-" * 60)
            return key_file, public_key
        else:
            # Generate public key from private key
            print("🔑 Generating public key from private key...")
            try:
                result = subprocess.run(
                    ["ssh-keygen", "-y", "-f", key_file],
                    capture_output=True,
                    text=True,
                    timeout=10
                )
                if result.returncode == 0:
                    public_key = result.stdout.strip()
                    with open(pub_key_file, 'w') as f:
                        f.write(public_key)
                    print(f"📋 Public key for Huawei Cloud import:")
                    print("-" * 60)
                    print(public_key)
                    print("-" * 60)
                    return key_file, public_key
                else:
                    print(f"❌ Failed to generate public key: {result.stderr}")
                    return None, None
            except Exception as e:
                print(f"❌ Error generating public key: {e}")
                return None, None
    else:
        print("🔑 Generating new SSH key pair...")
        try:
            result = subprocess.run(
                ["ssh-keygen", "-t", "rsa", "-b", "4096", "-f", key_file, "-N", "", "-q"],
                capture_output=True,
                text=True,
                timeout=30
            )
            
            if result.returncode == 0:
                print(f"✅ SSH key generated: {key_file}")
                
                # Read public key
                with open(f"{key_file}.pub", 'r') as f:
                    public_key = f.read().strip()
                
                print(f"📋 Public key for Huawei Cloud import:")
                print("-" * 60)
                print(public_key)
                print("-" * 60)
                return key_file, public_key
            else:
                print(f"❌ Failed to generate SSH key: {result.stderr}")
                return None, None
                
        except Exception as e:
            print(f"❌ Error generating SSH key: {e}")
            return None, None

def update_deployment_script(ip_address):
    """Update deployment script with correct IP address"""
    deploy_script = "deploy-app.sh"
    
    if not os.path.exists(deploy_script):
        print(f"❌ Deployment script not found: {deploy_script}")
        return False
    
    try:
        with open(deploy_script, 'r') as f:
            content = f.read()
        
        # Update IP address in script
        import re
        new_content = re.sub(r'IP="[^"]*"', f'IP="{ip_address}"', content)
        
        with open(deploy_script, 'w') as f:
            f.write(new_content)
        
        print(f"✅ Updated {deploy_script} with IP: {ip_address}")
        return True
        
    except Exception as e:
        print(f"❌ Failed to update deployment script: {e}")
        return False

def main():
    print("🔍 Checking Huawei Cloud ECS instance in Santiago region...")
    print("=" * 80)
    
    # Try to check via Huawei Cloud CLI
    print("📡 Checking via Huawei Cloud CLI...")
    instances = check_ecs_via_cli()
    
    if instances:
        print(f"✅ Found {len(instances)} ECS instances")
        for instance in instances:
            print(f"🔹 Instance: {instance.get('name', 'N/A')}")
            print(f"   ID: {instance.get('id', 'N/A')}")
            print(f"   Status: {instance.get('status', 'N/A')}")
            print(f"   Flavor: {instance.get('flavor', {}).get('id', 'N/A')}")
            
            # Check for floating IP
            addresses = instance.get('addresses', {})
            floating_ip = None
            for network in addresses.values():
                for addr in network:
                    if addr.get('OS-EXT-IPS:type') == 'floating':
                        floating_ip = addr.get('addr')
                        break
                if floating_ip:
                    break
            
            if floating_ip:
                print(f"   Public IP: {floating_ip}")
                
                # Check SSH connectivity
                print(f"🔌 Testing SSH connection to {floating_ip}...")
                if check_ecs_via_ssh(floating_ip):
                    print(f"✅ SSH connection successful!")
                    
                    # Generate SSH key
                    key_file, public_key = generate_ssh_key()
                    if key_file and public_key:
                        # Update deployment script
                        if update_deployment_script(floating_ip):
                            print(f"\n🎯 ECS instance is ready for deployment!")
                            print(f"   IP Address: {floating_ip}")
                            print(f"   SSH Key: {key_file}")
                            print(f"   Deployment script: ./deploy-app.sh")
                            print(f"\n🚀 To deploy PulseExpends:")
                            print(f"   1. Import the public key above to Huawei Cloud ECS Key Pairs")
                            print(f"   2. Bind the key to the ECS instance")
                            print(f"   3. Run: chmod +x deploy-app.sh")
                            print(f"   4. Run: ./deploy-app.sh")
                            return
                    else:
                        print("❌ Failed to generate SSH key")
                else:
                    print(f"❌ SSH connection failed")
                    print(f"   The instance might be stopped or SSH not configured")
                    print(f"   Please check the instance status in Huawei Cloud console")
            else:
                print("⚠️ No public IP found for this instance")
            
            print("-" * 80)
    else:
        print("❌ Could not retrieve ECS instances via CLI")
        print("\n📋 Alternative approach:")
        print("   1. Log into Huawei Cloud Console")
        print("   2. Go to ECS → Instances")
        print("   3. Check for instances in la-south-2 (Santiago) region")
        print("   4. Start any stopped instances")
        print("   5. Note the public IP address")
        print("\n🔑 Generate SSH key for deployment:")
        key_file, public_key = generate_ssh_key()
        if key_file and public_key:
            print(f"\n📋 Public key for Huawei Cloud import:")
            print("-" * 60)
            print(public_key)
            print("-" * 60)
            print(f"\n🔧 After getting instance IP, update deploy-app.sh:")
            print(f"   sed -i 's/IP=\"[^\"]*\"/IP=\"YOUR_INSTANCE_IP\"/' deploy-app.sh")
    
    print("\n💡 Manual deployment steps:")
    print("   1. Get ECS instance public IP from Huawei Cloud Console")
    print("   2. Generate SSH key if not already done")
    print("   3. Import public key to Huawei Cloud ECS Key Pairs")
    print("   4. Bind key to ECS instance")
    print("   5. Update deploy-app.sh with the IP address")
    print("   6. Run: chmod +x deploy-app.sh")
    print("   7. Run: ./deploy-app.sh")

if __name__ == "__main__":
    main()