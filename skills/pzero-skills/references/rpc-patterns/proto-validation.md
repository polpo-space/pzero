# Proto Field Validation

## Overview

pzero supports proto field validation through `protovalidate` and `buf/validate/validate.proto`.

## Common Patterns

```protobuf
string username = 1 [
  (buf.validate.field).string.min_len = 3,
  (buf.validate.field).string.max_len = 32
];

string email = 2 [
  (buf.validate.field).required = true,
  (buf.validate.field).string.email = true
];
```

Message-level validation:

```protobuf
option (buf.validate.message).cel = {
  id: "date_range.valid"
  message: "end_date must be after start_date"
  expression: "this.end_date > this.start_date"
};
```

## Guidance

- Prefer built-in constraints for simple rules
- Use CEL for complex rules
- Write descriptive validation messages
