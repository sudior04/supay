CREATE TABLE wallet_accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id uuid NOT NULL,
    account_type text NOT NULL DEFAULT 'PERSONAL' CHECK (account_type = 'PERSONAL'),
    currency char(3) NOT NULL DEFAULT 'VND' CHECK (currency = 'VND'),
    balance_minor bigint NOT NULL DEFAULT 0 CHECK (balance_minor >= 0),
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE','SUSPENDED','CLOSED')),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (owner_id, currency, account_type)
);

CREATE TABLE virtual_cards (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id uuid NOT NULL UNIQUE REFERENCES wallet_accounts(id) ON DELETE CASCADE,
    public_reference text NOT NULL UNIQUE,
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE','SUSPENDED','CLOSED')),
    created_at timestamptz NOT NULL DEFAULT now()
);
