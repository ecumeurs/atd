const Mocha = require('mocha');
const path = require('path');

const mocha = new Mocha({
  ui: 'tdd',
  color: true
});

mocha.addFile(path.join(__dirname, 'extension.test.js'));

mocha.run(failures => {
  process.exitCode = failures ? 1 : 0;
});