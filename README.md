# odoo-manifest

A Go parser for Odoo addon `__manifest__.py` files.

## Usage

```go
import odoomanifest "github.com/wimwenigerkind/odoo-manifest"

m, err := odoomanifest.Parse(data)
if err != nil {
    // handle malformed manifest
}
fmt.Println(m.Name, m.Version, m.Depends)
```

## Behaviour

It reads the Python dict literal without executing code: strings (single, double
and triple quoted, plus raw and unicode prefixes), numbers, `True`/`False`/`None`,
lists, tuples, dicts and set literals. It tolerates the coding line, comments,
trailing commas and implicit string concatenation, and rejects computed values
with a clear error.

Odoo defaults are applied when a key is absent: `license` defaults to `LGPL-3`,
`category` to `Uncategorized`, and `installable` to `true`. Version
interpretation is left to the caller.
