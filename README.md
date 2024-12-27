# Instashop E-commerce API

Instashop is a full-featured e-commerce platform providing a robust API for managing products, users, and products
It was written in GOLANG

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Admin Credentials](#admin-credentials)

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/hackdaemon2/instashop.git
   cd instashop-ecommerce-api
   ```

2. Install dependencies:

   ```sh
   go mod tidy
   ```

3. Set up environment variables:

   Copy the .env.sample file to .env and fill in the required values.

   ```sh
   cp .env.sample .env
   ```

## Configuration

  Configure the environment variables in the .env file:

   ```plaintext
   DB_HOST=localhost
   DB_PORT=3306
   DB_USER=root
   DB_PASSWORD=password
   DB_NAME=instashopdb
   ADMIN_PASSWORD=password
   ADMIN_EMAIL=john.doe@example.com
   ADMIN_FIRST_NAME=John
   ADMIN_LAST_NAME=Doe
   PORT=3000
   SECRET_KEY=secret
   ```

## Usage

Start the server:

   ```sh
   go run main.go
   ```

To Test the API open this link in your browser

   ```plaintext
   http://localhost:3000/swagger/index.html
   ```

## Admin Credentials

The admin account have been profiled with the following details:

- **Email:** john.doe@example.com
- **Password:** mayfay_2018@M1