# st2138_parameters Resource

Manages a reusable Catena/ST2138 parameter payload.

This resource defines reusable parameter payloads that can be passed into `st2138_device.parameters`.

## Example

```hcl
resource "st2138_parameters" "ooe_params" {
  parameters = [
    {
      counter        = 1
      number_example = 0
      string_example = "Hello World"
      float_array    = [1.1, 2.2, 3.3]
      struct_example = {
        nested_struct = {
          num_1 = 1
          num_2 = 2
        }
      }
    }
  ]
}

resource "st2138_parameters" "ooe_params_file" {
  parameters_file = "/path/to/file.stpm"
}
```

## Argument Reference

### Optional

- `parameters` (Dynamic): Parameter payload expressed as an object or list of objects.
- `parameters_file` (String): Path to a parameter file for bulk loading.

Constraint: At least one of `parameters` or `parameters_file` must be provided.

## Behavior

1. Stores a reusable parameter payload in Terraform state.
2. Allows other resources to reference a shared `parameters` value.
3. Supports file-path metadata through `parameters_file`.

## Notes

- When both attributes are set, `parameters` remains the authoritative payload for downstream references.
