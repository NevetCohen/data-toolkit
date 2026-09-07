---
name: data-toolkit-google-sheets
description: Bridge a Google Spreadsheet source or destination through the local Data Toolkit XLSX workflow. Use when the user provides a Google Sheets URL or ID, asks to deliver a Data Toolkit result as a Google Sheet, or asks to import/export a spreadsheet through the verified Google Drive connector.
---

# Data Toolkit Google Sheets bridge

Use Google Sheets only as a Codex-side XLSX bridge. Do not add it as a Go
`FormatID`, call network tools from the Data Toolkit runtime, use `clasp`, or
delete any Drive file. Read `docs/codex/workflow-json-format.md` before building
the local workflow template.

## Preconditions and stop rules

1. Detect whether the request has a Google Spreadsheet source, destination, or both.
2. Verify that the actual callable connector tools needed for the selected path
   are available: `google_drive_get_file_metadata`,
   `google_drive_export_file` for a source, `google_drive_import_spreadsheet`,
   `google_drive_search`, `google_drive_get_spreadsheet_metadata`,
   `google_drive_get_spreadsheet_cells`, and `google_drive_update_file` when a
   destination folder is requested.
3. If a required tool is unavailable, stop before inspection or local processing.
   Report `capability_gap`, advise the user to enable the Google Drive connector,
   and offer the manual fallback: export the source Sheet to XLSX for input, or
   accept a verified XLSX for manual upload after a local output run.
4. Stop with the same capability gap if native export exceeds the connector's
   10 MB response limit. Do not use a raw network download, `clasp`, or another
   connector as a workaround.
5. Never call a Drive deletion operation. Local cleanup may remove only the
   exact temporary XLSX paths created or returned in this run.

## Google Spreadsheet source

1. Call `google_drive_get_file_metadata` with the supplied file ID or URL.
   Verify that the result identifies the native Google Sheets MIME type
   `application/vnd.google-apps.spreadsheet`. Stop on an inaccessible, ambiguous,
   or non-Sheets file.
2. Call `google_drive_export_file` with the verified ID/URL and
   `mime_type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"`.
3. Read the successful result and use only its returned absolute `workspace_path`
   as the local source. Do not derive a path from title, ID, URL, or a guessed
   download directory.
4. Outside cleanup handling, pass that local XLSX through the normal sequence:
   capability discovery, inspection, workflow-template creation, `workflow
   validate --file`, and `run --file` with the expected digest.
5. In a finally-equivalent path, close local handles, verify the cleanup target
   is exactly the absolute `workspace_path` returned in step 3, and run
   `Remove-Item -LiteralPath '<workspace_path>' -Force`. Perform this cleanup
   whether the local workflow succeeds, fails, or is cancelled. Never delete
   the source Google Spreadsheet.

## Google Spreadsheet destination

1. Resolve the validated effective Google policy before upload:
   `google_sheets.destination_folder`, `google_sheets.title_template`, and
   `google_sheets.collision`. The default title template is `{output_stem}`.
   Style comes from `output.default_style` and `styles`, and is applied by the
   local Writer Engine before upload.
2. Run the normal local workflow. Continue only when Writer has published a
   valid XLSX after staging, reopen, format/schema validation, and assertions.
3. Resolve `destination_folder` to an exact folder ID (`root` is allowed), then
   call `google_drive_search` with `query: ""` and a
   `special_filter_query_str` equivalent to `name = '<escaped-title>' and
   mimeType = 'application/vnd.google-apps.spreadsheet' and
   '<folder-id>' in parents and trashed = false`. Follow every
   `next_page_token` with the same filter. Escape apostrophes in the title for
   Drive query syntax; do not use keyword matching as collision evidence.
4. Apply remote collision policy before import. Under `block`, stop if any
   exact match exists. Under `alternate_name`, test `<title> (2)`, then `(3)`,
   and so on, using the same exact paginated search, and select the first name
   with no match. Never overwrite a remote file.
5. Call `google_drive_import_spreadsheet` with the exact absolute local XLSX
   `source_file`, the resolved title, and `upload_mode: "native_google_sheets"`.
6. Verify the result reports conversion, native Google Sheets MIME type,
   Spreadsheet ID, and usable URL. Call `google_drive_get_spreadsheet_metadata`
   using that ID and verify the expected tabs and returned metadata.
7. When `destination_folder` is set, call `google_drive_get_file_metadata` for
   the imported file and read its current parents. Only then call
   `google_drive_update_file` with `addParents` set to the destination folder
   and `removeParents` set to the read-back current parent IDs. Read metadata
   again and verify the resulting parents.
8. Perform bounded structural and display readback with
   `google_drive_get_spreadsheet_cells` over explicit A1 ranges derived from the
   spreadsheet metadata; verify the requested headers, representative values,
   and display/style properties without scanning unbounded cells.
   Report the verified Spreadsheet URL together with the original local
   `RunReport` receipt.
9. In a finally-equivalent cleanup path, close local handles, verify the cleanup
   target is exactly the absolute local XLSX output created for this bridge, and
   run `Remove-Item -LiteralPath '<local-xlsx-output>' -Force`. Do not delete or
   overwrite the imported Drive file, including after an upload or readback
   failure.

## Required handoff records

Keep these values in the handoff/result: source or destination Drive ID, verified
MIME, returned absolute local path, selected local workflow file, workflow and
effective-config digests, local receipt ID, remote title/collision policy, and
the final verified Spreadsheet URL. Do not store credentials, connector tokens,
or a static operation catalog in the workflow file.
