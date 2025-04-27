#!/bin/bash
sudo apt-get update
sudo apt-get install -y docker.io docker-compose git

sudo usermod -aG docker azureuser

cd /home/azureuser
git clone --branch infra/dev https://github.com/SanjoDB/microservice-app-example.git
cd microservice-app-example

sudo docker-compose -f docker-compose.yml up -d --build