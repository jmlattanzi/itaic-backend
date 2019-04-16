# itaic-backend

Testing out a new version of the API for my previous project, ITAIC. This time I am experimenting with the Firestore, Redis, RabbitMQ, Docker.
This is the combo package for backend, it contains the database API and the cache API.

The compose file isn't working yet. I have to play with the timing since RabbitMQ takes a moment to spin up. In order to run this you need a few keys in a config. All you have to do is compile the images separately and run them in a cluster with containers for Redis and RabbitMQ. They are dependant on specific IP addresses.

This is still very early in development and is setup as such, so take all the code with a grain of salt.

## to-do:

1. Write tests
2. Finish cache enpoints
3. Finish Compose file
4. Build gateway
5. Build Jenkins pipeline (?)

## Structure

I am constantly tweaking the structure of this application, but for now the current architecture is laid out as such:

![alt text](https://s3-us-west-1.amazonaws.com/itaic/Service+diagram.png)
