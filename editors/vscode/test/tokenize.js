// 用 VS Code 自己用的 vscode-textmate + vscode-oniguruma 跑真实分词，
// 检查 token 是否落在预期的 scope 上。
//   node test/tokenize.js
// 带 -v 会打印每一行的 scope 表。

const fs = require('fs');
const path = require('path');
const vsctm = require('vscode-textmate');
const oniguruma = require('vscode-oniguruma');

const root = path.join(__dirname, '..');
const verbose = process.argv.includes('-v');
const grammarPath = path.join(root, 'syntaxes', 'unigo.tmLanguage.json');

const wasmBin = fs.readFileSync(
  path.join(root, 'node_modules', 'vscode-oniguruma', 'release', 'onig.wasm'),
).buffer;

let registry;
async function loadGrammar() {
  await oniguruma.loadWASM(wasmBin);
  const onigLib = {
    createOnigScanner: (patterns) => new oniguruma.OnigScanner(patterns),
    createOnigString: (s) => new oniguruma.OnigString(s),
  };
  registry = new vsctm.Registry({
    onigLib: Promise.resolve(onigLib),
    loadGrammar: async (scopeName) => {
      if (scopeName !== 'source.unigo') return null;
      return vsctm.parseRawGrammar(fs.readFileSync(grammarPath, 'utf8'), grammarPath);
    },
  });
}

// 每个断言用一段源码 + 期望落在某个 scope 上的文字。
const cases = [
  // 注释
  { line: '// 只有这一种注释', text: '//', scope: 'punctuation.definition.comment.unigo' },
  { line: '// 只有这一种注释', text: '只有这一种注释', scope: 'comment.line.double-slash.unigo' },
  { line: '// TODO: 修掉', text: 'TODO', scope: 'keyword.codetag.notation.unigo' },

  // package / import
  { line: 'package main;', text: 'package', scope: 'keyword.other.package.unigo' },
  { line: 'package main;', text: 'main', scope: 'entity.name.namespace.unigo' },
  { line: 'import fmt from "fmt";', text: 'fmt', scope: 'entity.name.namespace.unigo' },
  { line: 'import json from "encoding/json";', text: '"encoding/json"', scope: 'string.quoted.double.unigo' },

  // 字符串
  { line: 'var str1 :string = "这是字串符";', text: '这是字串符', scope: 'string.quoted.double.unigo' },
  { line: 'var str1 :string = "a\\nb";', text: '\\n', scope: 'constant.character.escape.unigo' },
  { line: 'var raw :string = `{"ok":true}`;', text: '{"ok":true}', scope: 'string.quoted.other.raw.unigo' },

  // 数字
  { line: 'var num :int = 520;', text: '520', scope: 'constant.numeric.integer.unigo' },
  { line: 'var ratio :float64 = 3.14;', text: '3.14', scope: 'constant.numeric.float.unigo' },
  { line: 'var big :float64 = 1.5e10;', text: '1.5e10', scope: 'constant.numeric.float.unigo' },

  // 变量声明与类型
  { line: 'var num :int = 520;', text: 'var', scope: 'storage.type.unigo' },
  { line: 'var num :int = 520;', text: 'num', scope: 'variable.other.declaration.unigo' },
  { line: 'var num :int = 520;', text: 'int', scope: 'support.type.unigo' },
  { line: 'const tStr :string = "hello";', text: 'tStr', scope: 'variable.other.declaration.unigo' },
  { line: 'var 世界 :string = "你好";', text: '世界', scope: 'variable.other.declaration.unigo' },
  { line: 'var (data :byte[], err :error) = os.ReadFile(path);', text: 'data', scope: 'variable.other.declaration.unigo' },
  { line: 'var (data :byte[], err :error) = os.ReadFile(path);', text: 'byte', scope: 'support.type.unigo' },
  { line: 'var scores :map[string, int] = {};', text: 'map', scope: 'storage.type.map.unigo' },
  { line: 'var item :Student = Student{};', text: 'Student', scope: 'entity.name.type.unigo' },
  { line: 'var raw :json.RawMessage = json.RawMessage(x);', text: 'json', scope: 'entity.name.namespace.unigo' },
  { line: 'var (file :*os.File, openErr :error) = os.Open(path);', text: '*', scope: 'keyword.operator.pointer.unigo' },

  // 函数声明与参数
  { line: 'func printStudent(student :string) :string {', text: 'func', scope: 'storage.type.function.unigo' },
  { line: 'func printStudent(student :string) :string {', text: 'printStudent', scope: 'entity.name.function.unigo' },
  { line: 'func printStudent(student :string) :string {', text: 'student', scope: 'variable.parameter.unigo' },
  { line: 'func load(path :string) :error {', text: 'load', scope: 'entity.name.function.unigo' },
  { line: 'func main() {', text: 'main', scope: 'entity.name.function.unigo' },

  // 结构体 / 接口 / 枚举 / 新类型
  { line: 'struct Student {', text: 'struct', scope: 'storage.type.struct.unigo' },
  { line: 'struct Student {', text: 'Student', scope: 'entity.name.type.unigo' },
  { line: 'struct Student {\n\tname :string;\n}', text: 'name', scope: 'variable.other.property.unigo' },
  { line: 'interface StudentParser {', text: 'StudentParser', scope: 'entity.name.type.unigo' },
  { line: 'enum Status {', text: 'Status', scope: 'entity.name.type.unigo' },
  { line: 'enum Status {\n\tready;\n\tfailed;\n}', text: 'failed', scope: 'variable.other.enummember.unigo' },
  { line: 'newtype Celsius = float64;', text: 'Celsius', scope: 'entity.name.type.unigo' },
  { line: 'newtype Celsius = float64;', text: 'float64', scope: 'support.type.unigo' },
  { line: 'alias Kelvin = float64;', text: 'alias', scope: 'storage.type.unigo' },

  // 控制流
  { line: 'if (ready) {', text: 'if', scope: 'keyword.control.unigo' },
  { line: '} elif (num > 0) {', text: 'elif', scope: 'keyword.control.unigo' },
  { line: '} else {', text: 'else', scope: 'keyword.control.unigo' },
  { line: 'while (count > 0) {', text: 'while', scope: 'keyword.control.loop.unigo' },
  { line: 'for (count > 0) {', text: 'for', scope: 'keyword.control.loop.unigo' },
  { line: 'switch (state) {', text: 'switch', scope: 'keyword.control.unigo' },
  { line: 'return item;', text: 'return', scope: 'keyword.control.flow.unigo' },
  { line: 'var (data :byte[], err :error) = os.ReadFile(path) else return err;', text: 'else', scope: 'keyword.control.unigo' },
  { line: 'var (data :byte[], err :error) = os.ReadFile(path) else return err;', text: 'return', scope: 'keyword.control.flow.unigo' },
  { line: 'var (score :int, found :bool) = scores["数学"];', text: 'bool', scope: 'support.type.unigo' },

  // case 与枚举成员
  { line: '\tcase Status.ready {', text: 'case', scope: 'keyword.control.unigo' },
  { line: '\tcase Status.ready {', text: 'Status', scope: 'entity.name.type.enum.unigo' },
  { line: '\tcase Status.ready {', text: 'ready', scope: 'variable.other.enummember.unigo' },
  { line: '\tdefault {', text: 'default', scope: 'keyword.control.unigo' },

  // 字面量与调用
  { line: 'var ready :bool = true;', text: 'true', scope: 'constant.language.boolean.unigo' },
  { line: 'return nil;', text: 'nil', scope: 'constant.language.null.unigo' },
  { line: 'var item :Student = Student{name = "树萌芽", age = 18};', text: 'Student', scope: 'entity.name.type.unigo' },
  { line: 'var item :Student = Student{name = "树萌芽", age = 18};', text: 'name', scope: 'variable.other.property.unigo' },
  { line: '\tvar written :int = fmt.Println(data) else return;', text: 'Println', scope: 'entity.name.function.member.unigo' },
  { line: '\tvar err :error = json.Unmarshal(raw, &item) else return Student{};', text: 'Unmarshal', scope: 'entity.name.function.member.unigo' },
  { line: 'var count :int = parse("520");', text: 'parse', scope: 'entity.name.function.unigo' },
  { line: '\tdelete(scores, "数学");', text: 'delete', scope: 'support.function.builtin.unigo' },
  { line: 'var parsed :Student = parseStudent(source);', text: 'source', scope: 'variable.other.unigo' },

  // 成员访问与运算符
  { line: 'var title :string = parsed.room.title;', text: 'room', scope: 'variable.other.property.unigo' },
  { line: 'var score :int = scores["语文"];', text: 'score', scope: 'variable.other.declaration.unigo' },
  { line: 'if (num > 0) {', text: '>', scope: 'keyword.operator.comparison.unigo' },
  { line: 'count = count - 1;', text: '-', scope: 'keyword.operator.arithmetic.unigo' },
  { line: 'if (ready && num > 0) {', text: '&&', scope: 'keyword.operator.logical.unigo' },
  { line: 'var err :error = json.Unmarshal(raw, &item) else return Student{};', text: '&', scope: 'keyword.operator.address.unigo' },

  // 语言禁止的写法要标红
  { line: 'data, err := os.ReadFile(path);', text: ':=', scope: 'invalid.illegal.short-assign.unigo' },
  { line: 'type Celsius float64;', text: 'type', scope: 'invalid.illegal.forbidden.unigo' },
  { line: 'fallthrough;', text: 'fallthrough', scope: 'invalid.illegal.forbidden.unigo' },
  { line: 'interface StudentParser {\n\tfunc parse(raw :string) :Student;\n}', text: 'parse', scope: 'entity.name.function.unigo' },
  { line: 'interface StudentParser {\n\tfunc parse(raw :string) :Student;\n}', text: 'Student', scope: 'entity.name.type.unigo' },
  { line: 'func main() {\n\tcount = count - 1;\n}', text: 'count', scope: 'variable.other.unigo' },
  { line: "'a'", text: "'a'", scope: 'invalid.illegal.character-literal.unigo' },
];

function tokenize(grammar, src) {
  const out = [];
  let stack = vsctm.INITIAL;
  for (const line of src.replace(/\r\n/g, '\n').split('\n')) {
    const r = grammar.tokenizeLine(line, stack);
    for (const t of r.tokens) out.push({ text: line.slice(t.startIndex, t.endIndex), scopes: t.scopes });
    stack = r.ruleStack;
  }
  return out;
}

async function main() {
  await loadGrammar();
  const grammar = await registry.loadGrammar('source.unigo');

  let pass = 0;
  const failures = [];
  for (const c of cases) {
    const tok = tokenize(grammar, c.line).find((t) => t.text.trim() === c.text);
    if (tok && tok.scopes.includes(c.scope)) pass++;
    else failures.push({ ...c, actual: tok ? tok.scopes.slice(1).join(' ') : '(文字没有单独成 token)' });
  }

  if (verbose) {
    const sample = fs.readFileSync(path.join(root, 'test', 'sample.ug'), 'utf8');
    for (const t of tokenize(grammar, sample)) {
      if (t.text.trim()) console.log(`${JSON.stringify(t.text).padEnd(28)} ${t.scopes.slice(1).join(' ')}`);
    }
  }

  console.log(`通过 ${pass}/${cases.length}`);
  if (failures.length) {
    console.log('\n失败:');
    for (const f of failures) {
      console.log(`  ${JSON.stringify(f.text)} 期望 ${f.scope}，实际 ${f.actual}`);
      console.log(`    源码: ${JSON.stringify(f.line)}`);
    }
    process.exitCode = 1;
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
