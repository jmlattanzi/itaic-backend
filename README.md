# itaic-backend

Testing out a new version of the API for my previous project, ITAIC. This time I am experimenting with the Firestore, Redis, RabbitMQ, Docker.
This is the combo package for backend, it contains the database API and the cache API.

In order to run this you need a few keys in a config. All you have to do is compile the images separately and run them in a cluster with containers for Redis and RabbitMQ. They are dependant on specific IP addresses.

This is still very early in development and is setup as such, so take all the code with a grain of salt.
