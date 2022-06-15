<div id="top"></div>

<br />
<div align="center">
  <a href="https://github.com/zaidallam/ZChainConceptBlockchain">
    <img src="images/logo.png" alt="Logo" width="80" height="80">
  </a>

  <h3 align="center">ZChain Concept Blockchain</h3>

  <p align="center">
    Concept blockchain with basic consensus functionality written in Go
    <br />
    <a href="https://github.com/zaidallam/ZChainConceptBlockchain"><strong>Explore the docs »</strong></a>
    <br />
  </p>
</div>

## About The Project

This is a very simple concept blockchain written in golang. The blockchain does not have its own currency, and instead simply stores arbitrary data, which can be virtually anything. Each running instance of the project acts as a ZChain node, and different nodes are simulated by running them on various localhost ports on your local machine.
Each node keeps a copy of the blockchain. Every node has the ability to propose a new block, and it waits on consensus between 80% or more of it's peer nodes before committing the proposed block to it's local record of the blockchain. Consensus is simple and is based on 80% or more of the peer nodes agreeing to add the block to their own local copy of the blockchain. In this fashion, all nodes' copies of the blockchain stay synchronized with eachother.
Nodes discover eachother by passing along lists of discovered nodes between one-another as they make requests and propose new blocks. Over time, this means that every node on the network will become aware of all the others.
The project uses a wide array of golang features and libraries, relative to it's size. No other major technologies are used.
To get started, please read below and follow the "Getting Started" directions

<p align="right">(<a href="#top">back to top</a>)</p>

### Built With

* [Go](https://go.dev/)
* [go-spew](github.com/davecgh/go-spew/spew)
* [Gorilla Toolkit's gorilla/mux](github.com/gorilla/mux)
* [GoDotEnv](github.com/joho/godotenv)

<p align="right">(<a href="#top">back to top</a>)</p>

## Prerequisites

This project requires you to have the Go programming language installed and configured on your system.

<p align="right">(<a href="#top">back to top</a>)</p>

## License

Distributed under the MIT License. See `LICENSE` for more information.

<p align="right">(<a href="#top">back to top</a>)</p>

## Contact

See the developer's portfolio, zaidallam.com, for contact info.

<p align="right">(<a href="#top">back to top</a>)</p>
