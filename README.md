# Floki
[![](https://dockerbuildbadges.quelltext.eu/status.svg?organization=niccokunzmann&repository=dockerhub-build-status-image)](https://hub.docker.com/r/nosinovacao/floki/builds/)  

> `tail -f > loki = floki`

Floki is a simple micro-service that aggregates [Kafka](https://kafka.apache.org/) logs, sorts them and pushes them to [Grafana Loki](https://github.com/grafana/loki). 

<!-- TABLE OF CONTENTS -->
## Table of Contents

* [About the Project](#about-the-project)
* [Getting Started](#getting-started)
  * [Usage](#usage)
* [Roadmap](#roadmap)
* [Contributing](#contributing)
* [License](#license)



<!-- ABOUT THE PROJECT -->
## About The Project
![Floki Diagram](images/floki.png)

[Loki](https://github.com/grafana/loki) has [Promtail](https://github.com/grafana/loki/tree/master/pkg/promtail) which is an agent that installs itself in every k8s cluster nodes and collects the pods logs and sends them to Loki. Since we use the [ELK Stack](https://www.elastic.co/pt/products/) in our infrastructure, each k8s node has the [Filebeat](https://www.elastic.co/pt/products/beats/filebeat) agent running for log collection, so by using Promtail, we are 
effectively duplicating the functionality. To avoid this, we have created Floki, which subscribes to our Kafka logging topics, orders the logs and sends them to Loki.

<!-- GETTING STARTED -->
## Getting Started
To run Floki, you need to have the [ELK Stack](https://www.elastic.co/pt/what-is/elk-stack) or other distributed logging system compatible with Kafka and also Loki installed in your infrastructure.

### Usage
```sh
docker run nosinovacao/floki \
   -lokiurl="http://<loki_base_url>/api/prom/push" \
   -brokerList="<kafka_broker_list>" \
   -topicPattern="^logging-*"
```

<!-- ROADMAP -->
## Roadmap

See the [open issues](https://github.com/nosinovacao/floki/issues) for a list of proposed features (and known issues).



<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request



<!-- LICENSE -->
## License

Distributed under the BSD-3-Clause License. See `LICENSE` for more information.



