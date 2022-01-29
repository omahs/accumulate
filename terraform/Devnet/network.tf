data "aws_availability_zones" "available" {}

resource "aws_vpc" "vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true #gives you an internal domain name
  enable_dns_hostnames = true #gives you an internal host name
  enable_classiclink   = false
  assign_generated_ipv6_cidr_block = true
  tags = {
    Name = "accumulate-devnet"
  }
}
resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.vpc.id
  tags = {
    Name = "accumulate-devnet"
  }
}

resource "aws_subnet" "subnet" {
  vpc_id                          = aws_vpc.vpc.id
  count                           = length(data.aws_availability_zones.available.names)
  cidr_block                      = cidrsubnet(aws_vpc.vpc.cidr_block, 8, count.index)
  availability_zone               = data.aws_availability_zones.available.names[count.index]
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.vpc.ipv6_cidr_block, 8, count.index)
  map_public_ip_on_launch         = true //it makes this a public subnet
  assign_ipv6_address_on_creation = true

  tags = {
    Name = "accumulate-devnet-${count.index}"
  }
}

resource "aws_route_table" "public" {
  vpc_id     = aws_vpc.vpc.id
  depends_on = [aws_internet_gateway.igw]

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  route {
    ipv6_cidr_block = "::/0"
    gateway_id      = aws_internet_gateway.igw.id
  }

  tags = {
    Name = "accumulate-devnet"
  }
}

resource "aws_route_table_association" "vpc" {
  count          = length(data.aws_availability_zones.available.names)
  subnet_id      = aws_subnet.subnet[count.index].id
  route_table_id = aws_route_table.public.id
}
