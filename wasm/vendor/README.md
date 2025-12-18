# Vendor assets

This directory contains the browser-side DXF preview dependencies so the demo
works in offline or restricted environments.

- `dxf-parser.js` (v1.1.2)
- `three.min.js` (v0.160.0)
- `three-dxf.js` (v1.3.1)

The accompanying `LICENSE.*` files are copied from each upstream package. To
update these assets, fetch the desired versions from npm (e.g. via `npm pack`),
then replace the dist/build outputs accordingly.
