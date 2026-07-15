const assert = require('assert');
const fs = require('fs');
const path = require('path');

describe('Extension Test Suite', () => {
  it('Test framework is working', () => {
    assert.ok(true, 'Mocha test framework is properly configured');
  });

  it('Extension structure is valid', () => {
    const extensionPath = path.join(__dirname, '..', 'extension.js');
    assert.ok(fs.existsSync(extensionPath), 'extension.js should exist');
  });
});