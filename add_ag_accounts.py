#!/usr/bin/env python3
"""Script to add accounts to MCP Antigravity server"""

import json
import subprocess
import sys

# Account data
accounts = [
    {
        "email": "yuoistarlk@gmail.com",
        "password": "Qwerty123.",
        "codeRecovery": "lgqbxji3hzb5c7bfivkzsg4vea3k5cld",
        "codes": [
            "9473-7712",
            "7289-6841",
            "7914-9548",
            "2555-8935",
            "2438-2271",
            "0659-5199",
            "4450-0953",
            "5933-7321",
            "2837-6601",
            "6111-9444"
        ]
    },
    {
        "email": "ddanangttiossawardi@gmail.com",
        "password": "mochicantik9",
        "codeRecovery": "vpkyti52aleb5ah6262mq2ieqemwfcco",
        "codes": [
            "7754-0065",
            "4006-7521",
            "2694-1892",
            "0445-0715",
            "6481-2169",
            "8595-9728",
            "0348-8157",
            "9009-1861",
            "4202-7296",
            "0423-5002"
        ]
    }
]

def send_jsonrpc_request(method, params):
    """Send JSON-RPC request to MCP server via stdio"""
    request = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": method,
        "params": params
    }
    
    # Convert to JSON string
    request_json = json.dumps(request) + "\n"
    
    # Run MCP server and send request
    try:
        process = subprocess.Popen(
            ["mcp-servers/bin/mcp-antigravity.exe"],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        
        # Send initialize request first
        init_request = {
            "jsonrpc": "2.0",
            "id": 0,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {},
                "clientInfo": {
                    "name": "add-accounts-script",
                    "version": "1.0.0"
                }
            }
        }
        process.stdin.write(json.dumps(init_request) + "\n")
        process.stdin.flush()
        
        # Read initialize response
        response = process.stdout.readline()
        print(f"Initialize response: {response}")
        
        # Send initialized notification
        process.stdin.write(json.dumps({"jsonrpc": "2.0", "method": "notifications/initialized"}) + "\n")
        process.stdin.flush()
        
        # Send actual request
        process.stdin.write(request_json)
        process.stdin.flush()
        
        # Read response
        response = process.stdout.readline()
        print(f"Response: {response}")
        
        process.stdin.close()
        process.wait(timeout=5)
        
        return json.loads(response)
    except Exception as e:
        print(f"Error: {e}")
        return None

def add_account_direct(account):
    """Add account using tools/call method"""
    params = {
        "name": "add_account",
        "arguments": {
            "email": account["email"],
            "password": account["password"],
            "codeRecovery": account["codeRecovery"],
            "codes": account["codes"]
        }
    }
    
    return send_jsonrpc_request("tools/call", params)

# Add accounts
print("Adding accounts to MCP Antigravity...")
for i, account in enumerate(accounts, 1):
    print(f"\n--- Adding Account {i}: {account['email']} ---")
    result = add_account_direct(account)
    if result:
        print(f"Success: {json.dumps(result, indent=2)}")
    else:
        print("Failed to add account")

print("\n✅ Done!")
