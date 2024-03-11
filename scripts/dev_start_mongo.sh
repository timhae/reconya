#! /bin/bash 

docker stop reconya-mongo-dev
docker start reconya-mongo-dev

# docker run --name reconya-mongo-dev -d -p 27017:27017 mongo:latest