# Example workflow lint tools

Configuration for [`examples.yml`](../workflows/examples.yml), which lints the
generated workflows in [`examples/`](/examples):

- `requirements.in` pins yamllint; `requirements.txt` adds hashes.
  After changing `requirements.in`, regenerate it with:

  ```shell
  uv pip compile --generate-hashes --universal --no-header \
    requirements.in -o requirements.txt
  ```

- `yamllint.yaml` is our configuration. `yamllint-curl.yaml` mirrors curl's,
  so their CI agrees with ours.
