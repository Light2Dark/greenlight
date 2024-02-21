# Greenlight

A tutorial on building APIs with Go. This tutorial is based on the book [Let's Go Further](https://lets-go-further.alexedwards.net/).

## Learnings (not exhaustive, just interesting)

- Enveloping responses - self-documenting responses
- Tradeoffs of formatting API responses

- Think about formatting APIs upfront and how we can maintain a consistent structure to API repossess

- Pointers vs values for receivers. For values, they can work on both pointers and values, while pointer methods only work on pointers.

- Overwriting structs
    - aux := struct{StructAlias: oldStruct, replaceField: val}

- Fail fast by panicking
    - handles unexpected errors and unhandled errors

- When passing untrusted data (eg. client to db sql), use placeholder params to prevent SQL injection attacks

- Pagination requires consistent ordering of rows returned. Do this by always ordering by primary key/unique col

- Recover function to recover from panics by catching the error :>