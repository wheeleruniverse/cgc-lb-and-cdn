# cgc-lb-and-cdn
https://acloudguru.com/blog/engineering/cloud-portfolio-challenge-load-balancing-and-content-delivery-network


The Cloud Portfolio Challenge focuses on building an image delivery service using foundational cloud components. This includes creating a service that returns at least one image matching a search criterion, using a minimum of two virtual machines (VMs) in the same region, serving images via a Content Delivery Network (CDN), and utilizing at least one load balancer as the internet entry point. A key requirement is that there should be no public access directly to the VMs. The challenge is cloud-agnostic, allowing you to use your preferred cloud platform, and requires submission via a GitHub repository with an architecture diagram and major decisions.

Here's what deployment on DigitalOcean would look like in terms of time, cost, complexity, and ease of development:

Time: DigitalOcean is known for its quick deployment times. You can deploy a new Droplet (their term for a VM) in less than a minute. The overall deployment time for the entire challenge would depend on your familiarity with cloud concepts and DigitalOcean's platform, but its developer-friendly nature suggests a relatively efficient setup process.

Cost: DigitalOcean offers predictable and generally affordable pricing.

Virtual Machines (Droplets): Basic Droplets start from approximately $4-$5 per month. Since the challenge requires at least two VMs, your base cost for compute would begin from around $8-$10 per month.

Load Balancers: Regional Load Balancers start at $12 per month per node. Global Load Balancers are available for $15 per month for basic usage, with costs increasing based on requests and data transfer.

Content Delivery Network (CDN) & Object Storage: DigitalOcean's Spaces object storage includes a built-in CDN, starting at $5 per month for 250 GB of storage and 1 TB of outbound transfer. Additional data transfer for the CDN is priced at $0.01 per GB.

Data Transfer: DigitalOcean offers free inbound data transfer, and Droplets include a free outbound data transfer allowance (starting at 500 GB per month, scaling up). Additional outbound transfer is billed at $0.01 per GB.

DigitalOcean also provides a free trial with a credit, which could be beneficial for initially setting up and testing the challenge environment.

Complexity: DigitalOcean is often praised for its simplicity and ease of use, particularly for developers. Its user interface and experience are designed to be intuitive and easy to navigate. The platform offers managed services for various components, which can significantly reduce the complexity of infrastructure management, allowing you to focus more on building your application.

Ease of Development: DigitalOcean's ecosystem is built with developers in mind, offering straightforward tools and managed services that simplify the deployment, scaling, and maintenance of applications. This focus on developer experience contributes to a smoother development process for challenges like the image delivery service.

For more detailed information on DigitalOcean's offerings and pricing, you can refer to their official documentation:

DigitalOcean Cloud Infrastructure for Developers

DigitalOcean Droplets Pricing

DigitalOcean Load Balancers Pricing

DigitalOcean Spaces Object Storage
