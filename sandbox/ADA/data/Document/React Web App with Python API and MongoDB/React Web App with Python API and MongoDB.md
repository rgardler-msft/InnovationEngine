# React Web App with Python API and MongoDB

## Overview

This architecture is designed to support a React-based web application that interacts with a Python-based API backend and uses MongoDB as the database. The architecture is ideal for modern web applications requiring dynamic user interfaces, scalable backend services, and flexible NoSQL database storage. Typical use cases include e-commerce platforms, social media applications, and content management systems. This document outlines the major components required to deploy this workload on Microsoft Azure.

### Major Components

* **Resource Group** - A logical container for all resources in this architecture. It simplifies management and billing by grouping related resources together.

* **Azure App Service for React Web App** - This service hosts the React web application, providing a scalable and managed platform for running front-end code. It ensures high availability and supports CI/CD pipelines for seamless deployment.

* **Azure App Service for Python API** - This service hosts the Python-based API backend, enabling RESTful communication between the front-end and database. It supports scaling based on demand and integrates with monitoring tools for performance tracking.

* **Azure Cosmos DB (MongoDB API)** - This resource provides a globally distributed, highly available NoSQL database that supports MongoDB APIs. It is used to store application data, such as user profiles, transactions, or content.

* **Azure Application Insights** - This service monitors the performance and health of both the React web app and Python API, providing actionable insights for debugging and optimization.

* **Azure Virtual Network (optional)** - If required, a virtual network can be used to securely connect resources, such as the API backend and database, ensuring private communication.

### Decision Points

In this architecture, users will encounter several decision points that can significantly impact the performance, cost, and scalability of the deployment. These include:

1. **App Service Plan SKU**: Users must choose the appropriate SKU for hosting the React web app and Python API. Factors include the expected traffic, performance requirements, and budget. For example, a Basic SKU may suffice for development, while a Premium SKU might be needed for production.

2. **Cosmos DB Configuration**: Decisions include the choice of throughput (manual or autoscale), global distribution, and indexing policies. These choices depend on the application's data access patterns and scalability needs.

3. **Application Insights Configuration**: Users must decide the level of monitoring required. For example, enabling advanced telemetry may increase costs but provide deeper insights into application performance.

4. **Networking**: If using a Virtual Network, users need to configure subnets, network security groups, and private endpoints. This is especially important for applications with strict security requirements.

5. **Scaling Strategy**: Both the App Services and Cosmos DB support autoscaling. Users must decide on scaling thresholds and limits based on anticipated workloads.

These decision points will be expanded upon later in the document to help users tailor the architecture to their specific needs.

### Alternatives

While this document focuses on deploying a React web app with a Python API and MongoDB on Azure, there are alternative approaches worth considering:

1. **Azure Kubernetes Service (AKS)**: For applications requiring container orchestration, AKS can be used to deploy the React app, Python API, and MongoDB containers. This approach offers greater control and flexibility but requires more management effort.

2. **Azure Functions**: For lightweight backend APIs, Azure Functions can replace the Python API App Service. This serverless option reduces costs for low-traffic applications but may not be suitable for complex APIs.

3. **Azure SQL Database**: If a relational database is preferred over MongoDB, Azure SQL Database can be used. It offers robust querying capabilities and is ideal for structured data.

These alternatives may be chosen based on specific requirements, such as the need for containerization, serverless architecture, or relational data storage. Each option has trade-offs in terms of complexity, cost, and scalability.

## Prerequisites

The following prerequisites are required before you are able to work through this document.

- Az CLI is installed and you are logged in to an active Azure subscription.

```bash
export SUFFIX=$(date +%s%N | sha256sum | head -c 6)
```

---

## Step 1: Create Resource Group

The first step is to create a resource group, which acts as a logical container for all resources in this architecture. Resource groups simplify management and billing by grouping related resources together.

Define the environment variable for the resource group name and region.

```bash
export RESOURCE_GROUP_NAME_ED345="ReactPythonMongoRG_$SUFFIX"
export REGION_ED345="westus2"
```

Create the resource group.

```bash
az group create --name $RESOURCE_GROUP_NAME_ED345 \
    --location $REGION_ED345
```

This command will output results similar to the following.

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/ReactPythonMongoRG_xxxxxx",
    "location": "westus2",
    "managedBy": null,
    "name": "ReactPythonMongoRG_xxxxxx",
    "properties": {
        "provisioningState": "Succeeded"
    },
    "tags": null,
    "type": "Microsoft.Resources/resourceGroups"
}
```

Resource groups are free to create, but the resources within them incur costs. Choose a region close to your user base to minimize latency and optimize performance.

---

## Step 2: Deploy Azure App Service for React Web App

This step deploys an Azure App Service to host the React web application. Azure App Service provides a scalable and managed platform for running front-end code.

Define environment variables for the App Service plan and web app.

```bash
export APP_SERVICE_PLAN_NAME_ED345="ReactAppPlan_$SUFFIX"
export WEB_APP_NAME_ED345="ReactWebApp$SUFFIX" # Removed underscore to comply with naming rules
export APP_SERVICE_SKU_ED345="P1V2" # Premium SKU for production workloads
```

Create the App Service plan.

```bash
az appservice plan create --name $APP_SERVICE_PLAN_NAME_ED345 \
    --resource-group $RESOURCE_GROUP_NAME_ED345 \
    --sku $APP_SERVICE_SKU_ED345 \
    --is-linux
```

Create the web app.

```bash
az webapp create --name $WEB_APP_NAME_ED345 \
    --resource-group $RESOURCE_GROUP_NAME_ED345 \
    --plan $APP_SERVICE_PLAN_NAME_ED345 \
    --runtime "NODE|16-lts"
```

This command will output results similar to the following.

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/ReactPythonMongoRG_xxxxxx/providers/Microsoft.Web/sites/ReactWebAppxxxxxx",
    "name": "ReactWebAppxxxxxx",
    "state": "Running",
    "hostNames": [
        "ReactWebAppxxxxxx.azurewebsites.net"
    ],
    "type": "Microsoft.Web/sites"
}
```

For production workloads, consider higher SKUs for better performance and scaling. Lower SKUs may suffice for development or testing environments.

---

## Step 3: Deploy Azure App Service for Python API

This step deploys an Azure App Service to host the Python-based API backend. This service enables RESTful communication between the front-end and database.

Define environment variables for the API App Service plan and web app.

```bash
export API_APP_SERVICE_PLAN_NAME_ED345="PythonAPIPlan_$SUFFIX"
export API_WEB_APP_NAME_ED345="PythonAPIApp$SUFFIX" # Removed underscore to comply with naming rules
export API_APP_SERVICE_SKU_ED345="P1V2" # Premium SKU for production workloads
```

Create the App Service plan.

```bash
az appservice plan create --name $API_APP_SERVICE_PLAN_NAME_ED345 \
    --resource-group $RESOURCE_GROUP_NAME_ED345 \
    --sku $API_APP_SERVICE_SKU_ED345 \
    --is-linux
```

Create the web app.

```bash
az webapp create --name $API_WEB_APP_NAME_ED345 \
    --resource-group $RESOURCE_GROUP_NAME_ED345 \
    --plan $API_APP_SERVICE_PLAN_NAME_ED345 \
    --runtime "PYTHON|3.9"
```

This command will output results similar to the following.

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/ReactPythonMongoRG_xxxxxx/providers/Microsoft.Web/sites/PythonAPIAppxxxxxx",
    "name": "PythonAPIAppxxxxxx",
    "state": "Running",
    "hostNames": [
        "PythonAPIAppxxxxxx.azurewebsites.net"
    ],
    "type": "Microsoft.Web/sites"
}
```

Ensure the runtime version matches your Python application requirements. Premium SKUs provide better scaling and performance for high-demand APIs.

---

## Step 4: Deploy Azure Cosmos DB (MongoDB API)

This step deploys Azure Cosmos DB configured with the MongoDB API to store application data.

Define environment variables for the Cosmos DB account.

```bash
export COSMOS_DB_ACCOUNT_NAME_ED345="mongodbacct$SUFFIX" # Adjusted name to comply with naming rules
export COSMOS_DB_THROUGHPUT_ED345="400" # Throughput in RU/s
```

Create the Cosmos DB account.

```bash
az cosmosdb create --name $COSMOS_DB_ACCOUNT_NAME_ED345 \
    --resource-group $RESOURCE_GROUP_NAME_ED345 \
    --kind MongoDB \
    --locations regionName=$REGION_ED345 failoverPriority=0 \
    --default-consistency-level "Session" \
    --enable-automatic-failover true
```

Configure throughput.

```bash
az cosmosdb mongodb database create --account-name $COSMOS_DB_ACCOUNT_NAME_ED345 \
    --resource-group $RESOURCE_GROUP_NAME_ED345 \
    --name "AppDatabase" \
    --throughput $COSMOS_DB_THROUGHPUT_ED345
```

This command will output results similar to the following.

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/ReactPythonMongoRG_xxxxxx/providers/Microsoft.DocumentDB/databaseAccounts/mongodbacct_xxxxxx",
    "name": "mongodbacct_xxxxxx",
    "type": "Microsoft.DocumentDB/databaseAccounts",
    "location": "westus2",
    "properties": {
        "consistencyPolicy": {
            "defaultConsistencyLevel": "Session"
        },
        "provisioningState": "Succeeded"
    }
}
```

Autoscale throughput can be enabled for dynamic workloads, but it may increase costs. Manual throughput is suitable for predictable workloads.

---

## Step 5: Enable Azure Application Insights

This step enables Azure Application Insights to monitor the performance and health of the React web app and Python API.

Define environment variables for Application Insights.

```bash
export APP_INSIGHTS_NAME_ED345="AppInsights_$SUFFIX"
```

Create Application Insights.

```bash
az monitor app-insights component create --app $APP_INSIGHTS_NAME_ED345 \
    --location $REGION_ED345 \
    --resource-group $RESOURCE_GROUP_NAME_ED345 \
    --application-type "web"
```

This command will output results similar to the following.

<!-- expected_similarity=0.3 -->

```text
{
    "id": "/subscriptions/xxxxx-xxxxx-xxxxx-xxxxx/resourceGroups/ReactPythonMongoRG_xxxxxx/providers/Microsoft.Insights/components/AppInsights_xxxxxx",
    "name": "AppInsights_xxxxxx",
    "type": "Microsoft.Insights/components",
    "location": "westus2",
    "properties": {
        "Application_Type": "web",
        "Flow_Type": "Bluefield",
        "Request_Source": "rest"
    }
}
```

Advanced telemetry options provide deeper insights but may increase costs. Basic monitoring is sufficient for most applications.

---

This concludes the deployment steps for the React-based web application architecture.

## Summary

This document detailed the architecture and step-by-step process for deploying a simple Python Flask application to Microsoft Azure using Azure App Service. The purpose of the architecture was to provide a lightweight, fully managed solution for hosting a Flask-based web application, ideal for small-scale projects, prototypes, or educational use cases. By leveraging Azure App Service, developers were able to focus on application development without the need to manage underlying infrastructure.

Key decision points included selecting an App Service Plan SKU, choosing an Azure region, configuring scaling options, determining deployment methods, and evaluating the need for persistent storage. The Free tier of the App Service Plan was recommended for demonstration purposes, while higher tiers were suggested for production workloads. Alternatives such as Azure Kubernetes Service, Azure Functions, and Virtual Machines were discussed for scenarios requiring more control, scalability, or customization.

### Next Steps

* **Test the Application** - Verify the functionality of the deployed Flask application by accessing its URL and ensuring it operates as expected.
* **Enable Monitoring** - Set up Azure Monitor or Application Insights to track performance, availability, and cost metrics for the deployed application.
* **Optimize for Production** - Evaluate scaling options, upgrade the App Service Plan tier, and integrate CI/CD pipelines for streamlined production deployments.

