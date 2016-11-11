
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE carts (
	id SERIAL PRIMARY KEY,
	access_token TEXT NOT NULL UNIQUE,
	is_checkout boolean DEFAULT 'false',

	created_at timestamp,
	updated_at timestamp
);
CREATE TABLE cart_items (
	id SERIAL PRIMARY KEY,
	cart_id SERIAL REFERENCES carts(id),
	product_id int NOT NULL,
	name varchar(255),
	price int NOT NULL DEFAULT 0,
	quantity int
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE cart_items;
DROP TABLE carts;
