### Snowflake

A trivial library for generating and managing globally-unique snowflake IDs(1). Snowflakes may be timestamp-based, or generated "semantically" (which is convenient when needing to map a table with an autoincrementing primary key to a globally-unique ID). Some helper functions and types are also included for serializing/deserializing IDs to strings, necessary when passing IDs to Javascript due to precision loss.

#### Importing
`go get github.com/cmertens/snowflake`


(1) https://instagram-engineering.com/sharding-ids-at-instagram-1cf5a71e5a5c
