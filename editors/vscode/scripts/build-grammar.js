// 生成 syntaxes/unigo.tmLanguage.json。
// 修改语法后运行: node scripts/build-grammar.js

const fs = require('fs');
const path = require('path');

// UniGo 的名字可以是中文等任意字母,所以用 Unicode 属性而不是 \w。
const B = '(?<![\\p{L}\\p{Nd}_])'; // 前边界
const E = '(?![\\p{L}\\p{Nd}_])'; // 后边界
const ID = '[\\p{L}_][\\p{L}\\p{Nd}_]*';
const kw = (words) => `${B}(?:${words.join('|')})${E}`;

// 类型：*os.File、map[string, int]、Student[]、json.RawMessage[][]
const QUAL = `${ID}(?:\\.${ID})?`;
const TYPE = `(?:\\*${QUAL}|map\\s*\\[[^\\]\\n]*\\](?:\\[\\])*|${QUAL}(?:\\[\\])*)`;

const BUILTIN_TYPES = [
  'bool', 'byte', 'rune', 'string', 'error', 'any',
  'int', 'int8', 'int16', 'int32', 'int64',
  'uint', 'uint8', 'uint16', 'uint32', 'uint64', 'uintptr',
  'float32', 'float64', 'complex64', 'complex128',
];
const BUILTIN_FUNCS = ['delete', 'len', 'cap', 'append', 'copy', 'panic', 'print', 'println', 'make', 'min', 'max', 'clear'];
const KEYWORDS = [
  'package', 'import', 'from', 'var', 'const', 'func', 'if', 'else', 'elif', 'while', 'for',
  'switch', 'case', 'default', 'enum', 'struct', 'interface', 'newtype', 'alias', 'return',
  'map', 'true', 'false', 'nil',
];
// 词法里保留但语言禁止使用的词。
const FORBIDDEN = ['fallthrough', 'iota', 'class', 'new', 'extends', 'implements', 'type'];

const typeCapture = { name: 'meta.type.unigo', patterns: [{ include: '#type-expression' }] };

// 名字 :类型
const annotated = (nameScope) => ({
  match: `${B}(${ID})\\s*(:)\\s*(${TYPE})`,
  captures: {
    1: { name: nameScope },
    2: { name: 'punctuation.separator.type.unigo' },
    3: typeCapture,
  },
});

const block = (keyword, storage, inner) => ({
  begin: `${B}(${keyword})${E}\\s+(${ID})\\s*(\\{)`,
  beginCaptures: {
    1: { name: storage },
    2: { name: 'entity.name.type.unigo' },
    3: { name: 'punctuation.section.braces.begin.unigo' },
  },
  end: '\\}',
  endCaptures: { 0: { name: 'punctuation.section.braces.end.unigo' } },
  name: `meta.${keyword}.unigo`,
  patterns: [{ include: '#comments' }, ...inner, { include: '#punctuation' }],
});

const grammar = {
  $schema: 'https://raw.githubusercontent.com/martinring/tmlanguage/master/tmlanguage.json',
  name: 'UniGo',
  scopeName: 'source.unigo',
  fileTypes: ['ug'],
  patterns: [{ include: '#statements' }],
  repository: {
    statements: {
      patterns: [
        { include: '#comments' },
        { include: '#strings' },
        { include: '#numbers' },
        { include: '#package-clause' },
        { include: '#import-declaration' },
        { include: '#struct-declaration' },
        { include: '#interface-declaration' },
        { include: '#enum-declaration' },
        { include: '#function-declaration' },
        { include: '#newtype-declaration' },
        { include: '#var-declaration' },
        { include: '#else-return' },
        { include: '#case-clause' },
        { include: '#keywords' },
        { include: '#annotated-name' },
        { include: '#composite-literal' },
        { include: '#field-initializer' },
        { include: '#calls' },
        { include: '#member-access' },
        { include: '#operators' },
        { include: '#punctuation' },
        { include: '#identifiers' },
      ],
    },

    comments: {
      patterns: [
        {
          begin: '//',
          beginCaptures: { 0: { name: 'punctuation.definition.comment.unigo' } },
          end: '$',
          name: 'comment.line.double-slash.unigo',
          patterns: [
            { match: `${B}(?:TODO|FIXME|XXX|NOTE|HACK)${E}`, name: 'keyword.codetag.notation.unigo' },
          ],
        },
      ],
    },

    strings: {
      patterns: [
        {
          begin: '"',
          beginCaptures: { 0: { name: 'punctuation.definition.string.begin.unigo' } },
          end: '"|$',
          endCaptures: { 0: { name: 'punctuation.definition.string.end.unigo' } },
          name: 'string.quoted.double.unigo',
          patterns: [
            { match: '\\\\[ntr\\\\"]', name: 'constant.character.escape.unigo' },
            { match: '\\\\.', name: 'invalid.illegal.unknown-escape.unigo' },
          ],
        },
        {
          begin: '`',
          beginCaptures: { 0: { name: 'punctuation.definition.string.begin.unigo' } },
          end: '`',
          endCaptures: { 0: { name: 'punctuation.definition.string.end.unigo' } },
          name: 'string.quoted.other.raw.unigo',
        },
        {
          // UniGo 不支持字符字面量。
          match: "'[^'\\n]*'?",
          name: 'invalid.illegal.character-literal.unigo',
        },
      ],
    },

    numbers: {
      patterns: [
        {
          match: `${B}[0-9]+(?:\\.[0-9]+(?:[eE][+-]?[0-9]+)?|[eE][+-]?[0-9]+)${E}`,
          name: 'constant.numeric.float.unigo',
        },
        { match: `${B}[0-9]+${E}`, name: 'constant.numeric.integer.unigo' },
      ],
    },

    'package-clause': {
      match: `${B}(package)${E}\\s+(${ID})`,
      captures: {
        1: { name: 'keyword.other.package.unigo' },
        2: { name: 'entity.name.namespace.unigo' },
      },
    },

    'import-declaration': {
      match: `${B}(import)${E}\\s+(${ID})\\s+(from)${E}\\s*("[^"\\n]*")`,
      captures: {
        1: { name: 'keyword.other.import.unigo' },
        2: { name: 'entity.name.namespace.unigo' },
        3: { name: 'keyword.other.from.unigo' },
        4: { name: 'string.quoted.double.unigo' },
      },
    },

    'struct-declaration': block('struct', 'storage.type.struct.unigo', [
      annotated('variable.other.property.unigo'),
    ]),

    'interface-declaration': block('interface', 'storage.type.interface.unigo', [
      { include: '#function-declaration' },
    ]),

    'enum-declaration': block('enum', 'storage.type.enum.unigo', [
      { match: `${B}${ID}${E}`, name: 'variable.other.enummember.unigo' },
    ]),

    // func 名字(参数) :返回类型
    'function-declaration': {
      begin: `${B}(func)${E}\\s+(${ID})\\s*(\\()`,
      beginCaptures: {
        1: { name: 'storage.type.function.unigo' },
        2: { name: 'entity.name.function.unigo' },
        3: { name: 'punctuation.definition.parameters.begin.unigo' },
      },
      end: `(\\))(?:\\s*(:)\\s*(${TYPE}))?`,
      endCaptures: {
        1: { name: 'punctuation.definition.parameters.end.unigo' },
        2: { name: 'punctuation.separator.type.unigo' },
        3: typeCapture,
      },
      name: 'meta.function.unigo',
      patterns: [
        { include: '#comments' },
        annotated('variable.parameter.unigo'),
        { match: ',', name: 'punctuation.separator.comma.unigo' },
      ],
    },

    'newtype-declaration': {
      match: `${B}(newtype|alias)${E}\\s+(${ID})\\s*(=)\\s*(${TYPE})`,
      captures: {
        1: { name: 'storage.type.unigo' },
        2: { name: 'entity.name.type.unigo' },
        3: { name: 'keyword.operator.assignment.unigo' },
        4: typeCapture,
      },
    },

    // var 名字 :类型 / const 名字 :类型 / var (a :T, b :U)
    'var-declaration': {
      patterns: [
        {
          match: `${B}(var|const)${E}\\s+(${ID})\\s*(:)\\s*(${TYPE})`,
          captures: {
            1: { name: 'storage.type.unigo' },
            2: { name: 'variable.other.declaration.unigo' },
            3: { name: 'punctuation.separator.type.unigo' },
            4: typeCapture,
          },
        },
        {
          begin: `${B}(var)${E}\\s*(\\()`,
          beginCaptures: {
            1: { name: 'storage.type.unigo' },
            2: { name: 'punctuation.section.parens.begin.unigo' },
          },
          end: '\\)',
          endCaptures: { 0: { name: 'punctuation.section.parens.end.unigo' } },
          name: 'meta.var.multiple.unigo',
          patterns: [
            annotated('variable.other.declaration.unigo'),
            { match: ',', name: 'punctuation.separator.comma.unigo' },
          ],
        },
      ],
    },

    'else-return': {
      match: `${B}(else)${E}\\s+(return)${E}`,
      captures: {
        1: { name: 'keyword.control.unigo' },
        2: { name: 'keyword.control.flow.unigo' },
      },
    },

    keywords: {
      patterns: [
        { match: kw(['if', 'elif', 'else', 'switch', 'case', 'default']), name: 'keyword.control.unigo' },
        { match: kw(['while', 'for']), name: 'keyword.control.loop.unigo' },
        { match: kw(['return']), name: 'keyword.control.flow.unigo' },
        { match: kw(['package', 'import', 'from']), name: 'keyword.other.import.unigo' },
        { match: kw(['var', 'const']), name: 'storage.type.unigo' },
        { match: kw(['func']), name: 'storage.type.function.unigo' },
        { match: kw(['enum', 'struct', 'interface', 'newtype', 'alias', 'map']), name: 'storage.type.unigo' },
        { match: kw(['true', 'false']), name: 'constant.language.boolean.unigo' },
        { match: kw(['nil']), name: 'constant.language.null.unigo' },
        { match: kw(FORBIDDEN), name: 'invalid.illegal.forbidden.unigo' },
      ],
    },

    // case Status.ready {
    // 不用变长后顾，VS Code 的 Oniguruma 不接受 (?<=case\s+)。
    'case-clause': {
      match: `${B}(case)${E}\\s+(${ID})\\s*(\\.)\\s*(${ID})`,
      captures: {
        1: { name: 'keyword.control.unigo' },
        2: { name: 'entity.name.type.enum.unigo' },
        3: { name: 'punctuation.accessor.unigo' },
        4: { name: 'variable.other.enummember.unigo' },
      },
    },

    'annotated-name': annotated('variable.other.unigo'),

    // Student{...} 或 StudentParser{ func ... }
    'composite-literal': {
      match: `${B}(?!(?:${KEYWORDS.join('|')})${E})(${ID})(?=\\{)`,
      captures: { 1: { name: 'entity.name.type.unigo' } },
    },

    // Student{name = "树萌芽"} 里的字段名。
    // TextMate 按行匹配，代码块的 { 后换行，块里的赋值不会被误判成字段。
    'field-initializer': {
      match: `([{,])\\s*(${ID})\\s*(?==(?!=))`,
      captures: {
        1: { name: 'punctuation.separator.unigo' },
        2: { name: 'variable.other.property.unigo' },
      },
    },

    calls: {
      patterns: [
        {
          match: `${B}(${BUILTIN_FUNCS.join('|')})\\s*(?=\\()`,
          captures: { 1: { name: 'support.function.builtin.unigo' } },
        },
        {
          // 类型转换，如 string(x)、json.RawMessage(x) 中的内置类型
          match: `${B}(${BUILTIN_TYPES.join('|')})\\s*(?=\\()`,
          captures: { 1: { name: 'support.type.unigo' } },
        },
        {
          // fmt.Println(...) 这类成员调用。不用后顾，避免变长断言。
          match: `(\\.)\\s*(${ID})\\s*(?=\\()`,
          captures: {
            1: { name: 'punctuation.accessor.unigo' },
            2: { name: 'entity.name.function.member.unigo' },
          },
        },
        {
          match: `${B}(?!(?:${KEYWORDS.join('|')})${E})(${ID})\\s*(?=\\()`,
          captures: { 1: { name: 'entity.name.function.unigo' } },
        },
      ],
    },

    'member-access': {
      match: `(\\.)\\s*(${ID})`,
      captures: {
        1: { name: 'punctuation.accessor.unigo' },
        2: { name: 'variable.other.property.unigo' },
      },
    },

    'type-expression': {
      patterns: [
        { match: kw(['map']), name: 'storage.type.map.unigo' },
        { match: kw(BUILTIN_TYPES), name: 'support.type.unigo' },
        {
          match: `(${ID})(\\.)(${ID})`,
          captures: {
            1: { name: 'entity.name.namespace.unigo' },
            2: { name: 'punctuation.accessor.unigo' },
            3: { name: 'entity.name.type.unigo' },
          },
        },
        { match: `${B}${ID}${E}`, name: 'entity.name.type.unigo' },
        { match: '\\*', name: 'keyword.operator.pointer.unigo' },
        { match: '\\[|\\]', name: 'punctuation.section.brackets.unigo' },
        { match: ',', name: 'punctuation.separator.comma.unigo' },
      ],
    },

    operators: {
      patterns: [
        { match: ':=', name: 'invalid.illegal.short-assign.unigo' },
        { match: '/\\*|\\*/', name: 'invalid.illegal.block-comment.unigo' },
        { match: '\\|\\||&&|!(?!=)', name: 'keyword.operator.logical.unigo' },
        { match: '==|!=|<=|>=|<|>', name: 'keyword.operator.comparison.unigo' },
        { match: '[+\\-*/%]', name: 'keyword.operator.arithmetic.unigo' },
        { match: '&', name: 'keyword.operator.address.unigo' },
        { match: '=', name: 'keyword.operator.assignment.unigo' },
      ],
    },

    punctuation: {
      patterns: [
        { match: '\\(', name: 'punctuation.section.parens.begin.unigo' },
        { match: '\\)', name: 'punctuation.section.parens.end.unigo' },
        { match: '\\{', name: 'punctuation.section.braces.begin.unigo' },
        { match: '\\}', name: 'punctuation.section.braces.end.unigo' },
        { match: '\\[', name: 'punctuation.section.brackets.begin.unigo' },
        { match: '\\]', name: 'punctuation.section.brackets.end.unigo' },
        { match: ',', name: 'punctuation.separator.comma.unigo' },
        { match: ';', name: 'punctuation.terminator.statement.unigo' },
        { match: ':', name: 'punctuation.separator.colon.unigo' },
        { match: '\\.', name: 'punctuation.accessor.unigo' },
      ],
    },

    identifiers: {
      match: `${B}${ID}${E}`,
      name: 'variable.other.unigo',
    },
  },
};

const out = path.join(__dirname, '..', 'syntaxes', 'unigo.tmLanguage.json');
fs.writeFileSync(out, JSON.stringify(grammar, null, 2) + '\n');
console.log('wrote', path.relative(process.cwd(), out));
