-- ORDERS
create table if not exists orders (
  id uuid primary key,
  user_id text not null,
  amount bigint not null check (amount > 0),
  description text not null check (char_length(description) <= 200),
  status text not null,
  created_at timestamptz not null default now()
);

create table if not exists orders_outbox (
  id bigserial primary key,
  message_id uuid not null unique,
  topic text not null,
  key text not null,
  payload jsonb not null,
  created_at timestamptz not null default now(),
  published_at timestamptz null
);

-- PAYMENTS
create table if not exists accounts (
  user_id text primary key,
  balance bigint not null check (balance >= 0)
);

create table if not exists payments_inbox (
  message_id uuid primary key,
  received_at timestamptz not null default now()
);

create table if not exists payments (
  order_id uuid primary key,
  user_id text not null,
  amount bigint not null check (amount > 0),
  status text not null,
  created_at timestamptz not null default now()
);

create table if not exists payments_outbox (
  id bigserial primary key,
  message_id uuid not null unique,
  topic text not null,
  key text not null,
  payload jsonb not null,
  created_at timestamptz not null default now(),
  published_at timestamptz null
);
