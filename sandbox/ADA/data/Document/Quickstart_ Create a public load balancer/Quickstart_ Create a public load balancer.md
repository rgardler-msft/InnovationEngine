# Quickstart: Create a public load balancer

## Overview

This document provides a step-by-step guide to creating a public load balancer in Azure using the Azure CLI. A public load balancer is a critical resource for distributing incoming internet traffic across multiple virtual machines (VMs) in a backend pool, ensuring high availability and scalability for applications. The guide includes instructions for deploying all necessary resources, such as virtual networks, subnets, public IP addresses, backend VMs, and additional components like Azure Bastion and NAT Gateway. This quickstart is ideal for users looking to set up a load balancer to manage internet traffic for their applications hosted on Azure.

### Major Components

* **Resource Group** - A logical container to manage and organize all resources deployed for the load balancer setup.
* **Virtual Network and Subnets** - Provides the networking infrastructure for the load balancer and backend VMs.
* **Public IP Address** - Enables internet-facing access to the load balancer.
* **Load Balancer** - Distributes incoming traffic across backend VMs, with components such as frontend IP pool, backend pool, health probes, and load balancing rules.
* **Network Security Group (NSG)** - Ensures secure inbound and outbound traffic for backend VMs.
* **Azure Bastion** - Provides secure access to backend VMs without exposing them to the internet.
* **Backend Virtual Machines** - Hosts the application workloads that the load balancer distributes traffic to.
* **NAT Gateway** - Provides outbound internet connectivity for backend resources.
* **IIS Installation** - Configures backend VMs with a web server to test the load balancer setup.

### Decision Points

In this architecture, users will need to make several decisions that impact performance, cost, and scalability. Key decision points include:

1. **Resource Group Location**: Choose the Azure region where resources will be deployed. Factors include proximity to users, compliance requirements, and cost.
2. **Virtual Network Configuration**: Define address spaces and subnet sizes based on anticipated traffic and resource requirements.
3. **Public IP Address SKU**: Select between Basic and Standard SKUs. Standard SKU supports zone redundancy and is recommended for production workloads.
4. **Load Balancer SKU**: Choose between Basic and Standard SKUs. Standard SKU offers advanced features like high availability zones and secure by default settings.
5. **Health Probe Configuration**: Determine the protocol (TCP/HTTP), port, and frequency for health checks to ensure backend VM availability.
6. **Load Balancer Rules**: Configure traffic distribution rules, including frontend/backend ports, protocol, and idle timeout settings.
7. **VM Size and Image**: Select VM sizes and operating system images based on application requirements and budget.
8. **NAT Gateway Idle Timeout**: Configure timeout settings for outbound connections based on application behavior.
9. **Azure Bastion Deployment**: Decide whether to deploy Azure Bastion for secure VM management, considering its cost and security benefits.

Each decision point involves trade-offs between cost, performance, and scalability. Additional details on these choices will be provided in the relevant sections of the document.

### Alternatives

While this document focuses on deploying a public load balancer using the Azure CLI, alternative approaches and configurations exist:

1. **Basic Load Balancer**: A simpler and lower-cost option for non-critical workloads. However, it lacks advanced features like zone redundancy and secure defaults.
2. **Application Gateway**: Provides Layer 7 (HTTP/HTTPS) load balancing with features like SSL termination, Web Application Firewall (WAF), and URL-based routing. Suitable for web applications requiring advanced traffic management.
3. **Azure Front Door**: A global load balancing solution for high-performance web applications. Offers features like caching, SSL offloading, and traffic acceleration.
4. **Azure Traffic Manager**: A DNS-based traffic routing solution for distributing traffic across multiple regions or endpoints.

Each alternative has specific use cases and trade-offs. For example, Application Gateway is ideal for web applications requiring Layer 7 routing, while Azure Front Door is suited for global-scale applications needing fast content delivery. The choice depends on the workload requirements and desired features.

## Prerequisites

The following prerequisites are required before you are able to work through this document.

- Az CLI is installed and you are logged in to an active Azure subscription.

```bash
export SUFFIX=$(date +%s%N | sha256sum | head -c 6)
```

---

## Step 1: Create a Resource Group

The resource group is a logical container that will hold all the resources for the load balancer setup. 

Define the following environment variables:

```bash
export RESOURCE_GROUP_NAME_ED42="LoadBalancerRG_$SUFFIX" # Name of the resource group
export REGION_ED42="westus2" # Azure region for deployment
```

Create the resource group:

```bash
az group create --name $RESOURCE_GROUP_NAME_ED42 \
    --location $REGION_ED42
```

This command will output results similar to the following:

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/LoadBalancerRG_xxxxxx",
    "location": "westus2",
    "managedBy": null,
    "name": "LoadBalancerRG_xxxxxx",
    "properties": {
        "provisioningState": "Succeeded"
    },
    "tags": null,
    "type": "Microsoft.Resources/resourceGroups"
}
```

Performance, cost, and reliability considerations: Resource groups do not incur costs directly, but they help organize resources for easier management. Choose a region close to your users for lower latency and compliance requirements.

---

## Step 2: Create a Virtual Network and Subnets

The virtual network provides the networking infrastructure for the load balancer and backend VMs. Subnets segment the network for better organization and security.

Define the following environment variables:

```bash
export VNET_NAME_ED42="LoadBalancerVNet_$SUFFIX" # Name of the virtual network
export VNET_ADDRESS_PREFIX_ED42="10.0.0.0/16" # Address space for the virtual network
export SUBNET_NAME_ED42="BackendSubnet_$SUFFIX" # Name of the subnet
export SUBNET_ADDRESS_PREFIX_ED42="10.0.1.0/24" # Address space for the subnet
```

Create the virtual network and subnet:

```bash
az network vnet create --resource-group $RESOURCE_GROUP_NAME_ED42 \
    --name $VNET_NAME_ED42 \
    --address-prefix $VNET_ADDRESS_PREFIX_ED42 \
    --subnet-name $SUBNET_NAME_ED42 \
    --subnet-prefix $SUBNET_ADDRESS_PREFIX_ED42
```

This command will output results similar to the following:

<!-- expected_similarity=0.3 -->

```text
{
    "newVNet": {
        "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/LoadBalancerRG_xxxxxx/providers/Microsoft.Network/virtualNetworks/LoadBalancerVNet_xxxxxx",
        "location": "westus2",
        "name": "LoadBalancerVNet_xxxxxx",
        "properties": {
            "addressSpace": {
                "addressPrefixes": [
                    "10.0.0.0/16"
                ]
            },
            "subnets": [
                {
                    "name": "BackendSubnet_xxxxxx",
                    "properties": {
                        "addressPrefix": "10.0.1.0/24"
                    }
                }
            ]
        }
    }
}
```

Performance, cost, and reliability considerations: Ensure the address space is large enough to accommodate future growth. Subnets help isolate resources for security and better traffic management.

---

## Step 3: Create a Public IP Address

The public IP address enables internet-facing access to the load balancer.

Define the following environment variables:

```bash
export PUBLIC_IP_NAME_ED42="LoadBalancerPublicIP_$SUFFIX" # Name of the public IP address
export PUBLIC_IP_SKU_ED42="Standard" # SKU for the public IP address (Basic or Standard)
```

Create the public IP address:

```bash
az network public-ip create --resource-group $RESOURCE_GROUP_NAME_ED42 \
    --name $PUBLIC_IP_NAME_ED42 \
    --sku $PUBLIC_IP_SKU_ED42 \
    --allocation-method Static
```

This command will output results similar to the following:

<!-- expected_similarity=0.3 -->

```text
{
    "publicIp": {
        "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/LoadBalancerRG_xxxxxx/providers/Microsoft.Network/publicIPAddresses/LoadBalancerPublicIP_xxxxxx",
        "location": "westus2",
        "name": "LoadBalancerPublicIP_xxxxxx",
        "properties": {
            "provisioningState": "Succeeded",
            "publicIpAllocationMethod": "Static",
            "sku": {
                "name": "Standard"
            }
        }
    }
}
```

Performance, cost, and reliability considerations: Standard SKU supports zone redundancy and is recommended for production workloads. Static allocation ensures the IP does not change, which is ideal for DNS configurations.

---

## Step 4: Create the Load Balancer

The load balancer distributes incoming traffic across backend VMs. This step configures the frontend IP pool, backend pool, health probes, and load balancing rules.

Define the following environment variables:

```bash
export LOAD_BALANCER_NAME_ED42="PublicLoadBalancer_$SUFFIX" # Name of the load balancer
```

Create the load balancer:

```bash
az network lb create --resource-group $RESOURCE_GROUP_NAME_ED42 \
    --name $LOAD_BALANCER_NAME_ED42 \
    --sku Standard \
    --frontend-ip-name FrontendIP_$SUFFIX \
    --public-ip-address $PUBLIC_IP_NAME_ED42 \
    --backend-pool-name BackendPool_$SUFFIX
```

This command will output results similar to the following:

<!-- expected_similarity=0.3 -->

```text
{
    "loadBalancer": {
        "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/LoadBalancerRG_xxxxxx/providers/Microsoft.Network/loadBalancers/PublicLoadBalancer_xxxxxx",
        "location": "westus2",
        "name": "PublicLoadBalancer_xxxxxx",
        "properties": {
            "frontendIPConfigurations": [
                {
                    "name": "FrontendIP_xxxxxx",
                    "properties": {
                        "publicIPAddress": {
                            "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/LoadBalancerRG_xxxxxx/providers/Microsoft.Network/publicIPAddresses/LoadBalancerPublicIP_xxxxxx"
                        }
                    }
                }
            ],
            "backendAddressPools": [
                {
                    "name": "BackendPool_xxxxxx"
                }
            ]
        }
    }
}
```

Performance, cost, and reliability considerations: Standard SKU provides advanced features like zone redundancy and secure defaults. Configure health probes and load balancing rules based on application requirements.

---

## Step 5: Create Backend Virtual Machines

The backend VMs host the application workloads distributed by the load balancer.

Define the following environment variables:

```bash
export VM_NAME_ED42="BackendVM_$SUFFIX" # Name of the virtual machine
export VM_SIZE_ED42="Standard_B1s" # Size of the virtual machine
export VM_IMAGE_ED42="Win2019Datacenter" # Operating system image
export ADMIN_USERNAME_ED42="azureuser" # Admin username for the VM
export ADMIN_PASSWORD_ED42="Password123!" # Admin password for the VM
```

Create the backend VM:

```bash
az vm create --resource-group $RESOURCE_GROUP_NAME_ED42 \
    --name $VM_NAME_ED42 \
    --image $VM_IMAGE_ED42 \
    --size $VM_SIZE_ED42 \
    --admin-username $ADMIN_USERNAME_ED42 \
    --admin-password $ADMIN_PASSWORD_ED42 \
    --vnet-name $VNET_NAME_ED42 \
    --subnet $SUBNET_NAME_ED42 \
    --nsg "" \
    --location $REGION_ED42
```

This command will output results similar to the following:

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/LoadBalancerRG_xxxxxx/providers/Microsoft.Compute/virtualMachines/BackendVM_xxxxxx",
    "location": "westus2",
    "name": "BackendVM_xxxxxx",
    "properties": {
        "provisioningState": "Succeeded"
    }
}
```

Performance, cost, and reliability considerations: Choose VM sizes based on workload requirements. Smaller sizes like `Standard_B1s` are cost-effective for testing, while larger sizes may be needed for production workloads.

---

## Step 6: Install IIS on Backend VM

Install IIS to configure the backend VM with a web server for testing the load balancer setup.

Run the following script on the backend VM using Azure CLI or remote access:

```bash
az vm run-command invoke --resource-group $RESOURCE_GROUP_NAME_ED42 \
    --name $VM_NAME_ED42 \
    --command-id RunPowerShellScript \
    --scripts "Install-WindowsFeature -name Web-Server -IncludeManagementTools"
```

This command will output results similar to the following:

<!-- expected_similarity=0.3 -->

```text
{
    "value": [
        {
            "code": "ComponentStatus/StdOut/succeeded",
            "level": "Info",
            "displayStatus": "Provisioning succeeded",
            "message": "Success: IIS has been installed."
        }
    ]
}
```

Performance, cost, and reliability considerations: IIS installation is lightweight and suitable for testing. For production, ensure the VM is configured with necessary security and performance optimizations.

---

## Step 7: Create Azure Bastion

Azure Bastion provides secure access to backend VMs without exposing them to the internet.

Define the following environment variables:

```bash
export BASTION_NAME_ED42="BastionHost_$SUFFIX" # Name of the Azure Bastion host
```

Create Azure Bastion:

```bash
az network bastion create --resource-group $RESOURCE_GROUP_NAME_ED42 \
    --name $BASTION_NAME_ED42 \
    --vnet-name $VNET_NAME_ED42 \
    --location $REGION_ED42
```

This command will output results similar to the following:

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/LoadBalancerRG_xxxxxx/providers/Microsoft.Network/bastionHosts/BastionHost_xxxxxx",
    "location": "westus2",
    "name": "BastionHost_xxxxxx",
    "properties": {
        "provisioningState": "Succeeded"
    }
}
```

Performance, cost, and reliability considerations: Azure Bastion improves security by eliminating the need for public IPs on VMs. It incurs additional costs but is highly recommended for production environments.

---

## Step 8: Create NAT Gateway

The NAT Gateway provides outbound internet connectivity for backend resources.

Define the following environment variables:

```bash
export NAT_GATEWAY_NAME_ED42="NATGateway_$SUFFIX" # Name of the NAT Gateway
```

Create the NAT Gateway:

```bash
az network nat gateway create --resource-group $RESOURCE_GROUP_NAME_ED42 \
    --name $NAT_GATEWAY_NAME_ED42 \
    --public-ip-addresses $PUBLIC_IP_NAME_ED42 \
    --location $REGION_ED42
```

This command will output results similar to the following:

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/LoadBalancerRG_xxxxxx/providers/Microsoft.Network/natGateways/NATGateway_xxxxxx",
    "location": "westus2",
    "name": "NATGateway_xxxxxx",
    "properties": {
        "provisioningState": "Succeeded"
    }
}
```

Performance, cost, and reliability considerations: NAT Gateway ensures reliable outbound connectivity and reduces idle timeout issues. It incurs additional costs but is essential for production workloads.

---

This completes the deployment of the public load balancer architecture.

## Summary

This document detailed the architecture and step-by-step process for deploying a simple Python Flask application to Microsoft Azure using Azure App Service. The purpose of the architecture was to provide a lightweight, fully managed solution for hosting a Flask-based web application, ideal for small-scale projects, prototypes, or educational use cases. By leveraging Azure App Service, developers were able to focus on application development without the need to manage underlying infrastructure.

Key decision points included selecting an App Service Plan SKU, choosing an Azure region, configuring scaling options, determining deployment methods, and evaluating the need for persistent storage. The Free tier of the App Service Plan was recommended for demonstration purposes, while higher tiers were suggested for production workloads. Alternatives such as Azure Kubernetes Service, Azure Functions, and Virtual Machines were discussed for scenarios requiring more control, scalability, or customization.

### Next Steps

* **Test the Application** - Verify the functionality of the deployed Flask application by accessing its URL and ensuring it operates as expected.
* **Enable Monitoring** - Set up Azure Monitor or Application Insights to track performance, availability, and cost metrics for the deployed application.
* **Optimize for Production** - Evaluate scaling options, upgrade the App Service Plan tier, and integrate CI/CD pipelines for streamlined production deployments.

