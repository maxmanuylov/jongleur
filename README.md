# jongleur
Simple dynamic etcd-backed L4 TCP load balancer.

One need just two simple steps to get jongleur ready (the order of the steps does not matter).

1. Start "jongleur item" daemon for every instance of the load-balanced service:

   ```sh
   jongleur item --type=my-service --host=1.2.3.4:5678 [--health=http://1.2.3.4:999/healthStatus] [--etcd=http://127.0.0.1:2379]
   ```
   
   It makes sense to start this daemon on the machine where your service instance runs though it is not obligatory.
   Run `jongleur item --help` for more detailed options description.
   
2. Start "jongleur" daemon to load-balance the service instances:
   
   ```sh
   jongleur --items=my-service --listen=:1234 [--etcd=http://127.0.0.1:2379]
   ```
   
   This daemon runs a proxy on `tcp://0.0.0.0:1234` that load-balances all the requests among the service instances.
   
   It makes sense to have jongleur proxy locally on every machine from which you want to access your service rather than having just one centralized proxy.
   Run `jongleur --help` for more detailed options description.
   
## How it works

Every jongleur item creates a TTL'ed etcd key describing its service instance. Then it periodically checks for the service instance health status and refreshes the TTL.
Every jongleur proxy watches etcd for the changes and dynamically updates own list of the service instances.

The only important thing you need to do is to specify the same item type for both "jongleur item" and "jongleur" daemons ("my-service" in the example above). All the rest is handled automatically.
