// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

/* eslint-disable import/no-commonjs */

module.exports = {
  process (src, filename) {
    return `module.exports = "${filename}";`
  },
}