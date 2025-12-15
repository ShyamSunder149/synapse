# Synapse Framework

## Overview

Synapse is a modular and configurable framework to build unified web crawling and data processing pipelines in Go for any workload from single machine to distributed clusters (planned for the future) and provides core components including fetcher, frontier, pipeline, queue and storage backends, allowing the developers focus on domain-specific crawling logic without re-implementing infrastructure from scratch, rather configure and extend the existing components.

## Status

This framework is in active development and not production-ready yet. Breaking changes may occur in future releases. So, the public documentation and examples will be provided once the core components stabilize.

## Documentation

For developers, component-specific implementation details are available in their respective directories with examples ([Fetcher](./fetcher), [Spooler](./spooler))

## Contributing

Contributions are welcome! Please refer to the [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines on how to contribute to this project.

## History and Purpose

This framework is primarily developed as an engineering challenge to develop highly pluggable distributed web crawler infrastructure in Go.

There is a long-term goal to leverage crawling infrastructure (on top of this framework) in a separate [OSS search engine project](https://github.com/ritvikos/idx), that project remains entirely independent.

All architectural decisions, design choices, and implementation details in Synapse are made solely based on its merit as a standalone Go web crawling framework, with no influence from or coupling to any future projects.

## Ethical Considerations

It's not intended for any malicious or unethical web scraping/crawling activities. Please ensure you [comply with the website's `robots.txt` directives](./frontier/robots) and terms of service (TOS) before crawling/scraping.
