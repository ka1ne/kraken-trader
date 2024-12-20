# Kraken Trader

A CLI tool for automated trading on Kraken exchange.

## Installation

Clone the repository

```bash
git clone https://github.com/ka1ne/kraken-trader
cd kraken-trader
```

Install dependencies

```bash
go mod download
```

Build

```bash
go build -o kraken-trader
```

## Configuration

Create a `.env` file in the project root:

```bash
KRAKEN_API_KEY=your_api_key_here
KRAKEN_API_SECRET=your_api_secret_here
```

## Usage

### Place a Limit Order

#### Buy 0.002 ETH at $1000

```bash
./kraken-trader place-order --pair ETH/USD --side buy --price 1000 --volume 0.002
```

#### Sell 0.1 BTC at $200000

```bash
./kraken-trader order --pair BTC/USD --side sell --price 200000 --volume 0.1
```

### Place a Market Order

#### Buy 0.002 ETH at market price

```bash
./kraken-trader order --pair ETH/USD --side buy --volume 0.002
```

### Trailing Entry Orders

#### Buy when price enters $45000-$50000 range

```bash
./kraken-trader trailing --pair BTC/USD --side buy --upper 50000 --lower 45000 --volume 0.01 --orders 5
```

#### Sell when price enters $45000-$50000 range

```bash
./kraken-trader trailing --pair BTC/USD --side sell --upper 50000 --lower 45000 --volume 0.01 --orders 5
```

## Development

### Install tools

```bash
make install-tools
```


### Run tests

```bash
make test
```

### Run integration tests

```bash
make test-integration
```

### Generate test coverage

```bash
make test-coverage
```