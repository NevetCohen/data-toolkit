# Data Toolkit workflow JSON format

Use this instruction before creating, editing, validating, or running a Data
Toolkit V1 workflow file. The authoritative runtime schemas still come from
`capabilities --json` and `describe --json`; this document defines the stable
file, variable, and CLI procedure.

## Template document

Every file is one strict `WorkflowTemplateDocument`:

```json
{
  "$schema": "urn:data-toolkit:workflow-template:v1",
  "schema_version": "v1",
  "variables": [
    {
      "id": 1,
      "name": "source_path",
      "type": "string",
      "required": true,
      "description": "Absolute path to the input workbook"
    },
    {
      "id": 2,
      "name": "output_path",
      "type": "string",
      "required": true
    },
    {
      "id": 3,
      "name": "city",
      "type": "string",
      "required": false,
      "default": "בת ים"
    },
    {
      "id": 4,
      "name": "city_column_id",
      "type": "string",
      "required": true,
      "description": "Stable column ID returned by inspection"
    }
  ],
  "workflow": {
    "schema_version": "v1",
    "id": "city-filter",
    "source": {
      "id": "source-1",
      "path": { "$var": 1 },
      "format_id": "excel"
    },
    "operations": [
      {
        "id": "filter-city",
        "operation_id": "row.filter",
        "version": "v1",
        "parameters": {
          "expression": {
            "op": "eq",
            "column_id": { "$var": 4 },
            "value": { "type_id": "string", "value": { "$var": 3 } }
          }
        }
      }
    ],
    "output": {
      "path": { "$var": 2 },
      "format_id": "excel"
    },
    "validations": [
      {
        "kind": "all_rows_match",
        "expression": {
          "op": "eq",
          "column_id": { "$var": 4 },
          "value": { "type_id": "string", "value": { "$var": 3 } }
        }
      }
    ]
  }
}
```

Fill `workflow` only with fields accepted by the current strict
`WorkflowRequest` schema. Resolve source details through inspection and obtain
operation parameters from `describe`; do not invent catalog entries or schemas.

## Variables

- `id` is a positive integer. IDs are contiguous and start at `1`.
- `name` is descriptive and unique. It is documentation, not a substitution
  target.
- `type` is the declared runtime type. `required` is always explicit. `default`
  and `description` are optional.
- A reference is the complete value object `{"$var": 1}`. It may replace a
  JSON value only; never embed it in a string, a field name, or a JSON key.
- A required variable without a value or default stops validation. An unknown
  ID, non-contiguous ID, wrong type, or reference in an invalid position stops
  validation.
- Substitution happens before strict workflow validation and digest calculation.
  The same template, values, configuration, and runtime descriptors must yield
  the same workflow digest.

## CLI procedure

Validate a template before running it. Always request JSON output for agent use.

```powershell
data-toolkit workflow validate --file .\city-filter.json --var '1=C:\data\input.xlsx' --var '2=C:\data\bat-yam.xlsx' --var '4=city-column-id' --json
```

Copy the returned workflow digest exactly into the run call:

```powershell
data-toolkit run --file .\city-filter.json --var '1=C:\data\input.xlsx' --var '2=C:\data\bat-yam.xlsx' --var '4=city-column-id' --expected-workflow-digest '<digest>' --json
```

`--var` is repeatable and uses `<id>=<value>`. Supply a value that matches its
declared type. For a string, the text after `=` is the string value. For number,
boolean, null, array, or object variables use valid JSON representation.

`--config` loads user configuration. `--config-set` is repeatable and uses
`<json-pointer>=<json-value>`; its value is always JSON:

```powershell
data-toolkit workflow validate --file .\city-filter.json --config .\user-config.json --config-set '/runtime/maximum_memory_bytes=268435456' --config-set '/display/date_format="02/01/06"' --json
```

In PowerShell, single-quote the complete `--var` or `--config-set` argument
when it contains backslashes, spaces, quotes, braces, or JSON punctuation. Do
not rely on unquoted shell interpolation.

Configuration precedence is: defaults, user configuration, CLI patches,
workflow overrides, output overrides. The resulting configuration is fully
validated and included in the effective-configuration digest.

## Config patch limits

Reject duplicate paths, missing paths, structural changes that make the config
invalid, and any patch targeting a protected path, its parent, or its child.
The protected list is `cli.non_overridable_config_paths`; it protects at least
`/schema_version`, the list itself, `/report/receipt_store`, and
`/runtime/temporary_workspace`.

## Stop conditions

Stop and report the structured error instead of guessing when:

- inspection leaves sheet, header, delimiter, root path, or output target
  ambiguous;
- no runtime operation expresses the requested result;
- template, variable, config patch, workflow validation, or expected digest
  fails;
- source and output normalize to the same path;
- a dirty-input, collision, resource-limit, or assertion decision would change
  the user's requested result.

For a Google Spreadsheet source or target, hand off to
`data-toolkit-google-sheets` before local processing. That Skill checks the
actual Google Drive connector and performs the XLSX bridge; do not model Google
Sheets as a workflow format.
