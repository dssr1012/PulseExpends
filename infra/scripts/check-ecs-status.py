#!/usr/bin/env python3
import json
from huaweicloudsdkcore.auth.credentials import BasicCredentials
from huaweicloudsdkecs.v2.region.ecs_region import EcsRegion
from huaweicloudsdkecs.v2 import *
from huaweicloudsdkcore.exceptions import exceptions

# Read credentials from terraform.tfvars
with open('terraform.tfvars', 'r') as f:
    content = f.read()
    
# Simple parsing of terraform.tfvars
import re
access_key = re.search(r'access_key\s*=\s*"([^"]+)"', content).group(1)
secret_key = re.search(r'secret_key\s*=\s*"([^"]+)"', content).group(1)
project_id = re.search(r'project_id\s*=\s*"([^"]+)"', content).group(1)
region = re.search(r'region\s*=\s*"([^"]+)"', content).group(1)

print(f"Using credentials:")
print(f"  Access Key: {access_key[:10]}...")
print(f"  Secret Key: {secret_key[:10]}...")
print(f"  Project ID: {project_id}")
print(f"  Region: {region}")
print()

# Create credentials
credentials = BasicCredentials(access_key, secret_key, project_id)

# Create client
client = EcsClient.new_builder() \
    .with_credentials(credentials) \
    .with_region(EcsRegion.value_of(region)) \
    .build()

try:
    # List ECS instances
    request = ListServersDetailsRequest()
    response = client.list_servers_details(request)
    
    print(f"Found {len(response.servers)} ECS instances:")
    print()
    
    for server in response.servers:
        print(f"Instance: {server.name}")
        print(f"  ID: {server.id}")
        print(f"  Status: {server.status}")
        
        # Get power state
        if hasattr(server, 'OS-EXT-STS:power_state'):
            power_state = getattr(server, 'OS-EXT-STS:power_state')
            power_states = {
                0: 'NOSTATE',
                1: 'RUNNING',
                3: 'PAUSED',
                4: 'SHUTDOWN',
                6: 'CRASHED',
                7: 'SUSPENDED'
            }
            print(f"  Power State: {power_states.get(power_state, f'UNKNOWN ({power_state})')}")
        
        # Get task state
        if hasattr(server, 'OS-EXT-STS:task_state'):
            task_state = getattr(server, 'OS-EXT-STS:task_state')
            print(f"  Task State: {task_state}")
        
        # Get VM state
        if hasattr(server, 'OS-EXT-STS:vm_state'):
            vm_state = getattr(server, 'OS-EXT-STS:vm_state')
            print(f"  VM State: {vm_state}")
        
        # Get addresses
        if server.addresses:
            print(f"  Addresses:")
            for network_name, addresses in server.addresses.items():
                for addr in addresses:
                    print(f"    {network_name}: {addr.addr} ({addr.version})")
        
        # Get key pair
        if server.key_name:
            print(f"  Key Pair: {server.key_name}")
        
        # Get created time
        if server.created:
            print(f"  Created: {server.created}")
        
        # Get updated time
        if server.updated:
            print(f"  Updated: {server.updated}")
        
        # Get flavor
        if hasattr(server, 'flavor'):
            flavor = server.flavor
            print(f"  Flavor: {flavor.id if hasattr(flavor, 'id') else 'N/A'}")
        
        # Get image
        if hasattr(server, 'image'):
            image = server.image
            print(f"  Image: {image.id if hasattr(image, 'id') else 'N/A'}")
        
        print("-" * 50)
        
except exceptions.ClientRequestException as e:
    print(f"Error: {e}")
    print(f"Error code: {e.error_code}")
    print(f"Error message: {e.error_msg}")
except Exception as e:
    print(f"Unexpected error: {e}")