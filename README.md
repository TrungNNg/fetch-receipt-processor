# Fetch Rewards Receipt Processor Challenge

---

## **Description**
This project is a web service implementation for the provided API as described in the [Fetch Rewards Receipt Processor Challenge](https://github.com/fetch-rewards/receipt-processor-challenge).

---

## **Usage**

### **1. Clone the repository**
Run the following command to clone the repository to your local machine:

```bash
git clone https://github.com/TrungNNg/fetch-receipt-processor.git
cd fetch-receipt-processor
```

### **2. Build the Docker Image**
```bash
docker build . -t fetch:latest
```

### **3. Run the Docker container**
```bash
docker run -p 8080:4000 fetch
```
By default, the server will listen on port `4000` inside the container, but it will be accessible on `localhost:8080` due to the port mapping.

Once the container is running, you can access the API at: [http://localhost:8080](http://localhost:8080)

---