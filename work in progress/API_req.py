import uuid
import json
import requests
# Generate a unique identifier
cluster_uid = str(uuid.uuid4())

# Create the cluster definition with the UID
cluster_definition = {
    "cluster": {
        "name": "clusterName",
        "platformType": "Docker",
        "agentsList": [
            "agent1"
        ],
        "domainName": "example.org",
        "creationTime": "Feb 08 2023 21:02:10",
        "managedBy": "",
        "uid": cluster_uid  # Insert the generated UID here
    }
}

# Convert to JSON string if needed
json_data = json.dumps(cluster_definition)

# Print or send json_data to your API
print(json_data)
url = "http://your.api.endpoint"
response = requests.post(url, json=cluster_definition)

print(response.status_code)
print(response.json())