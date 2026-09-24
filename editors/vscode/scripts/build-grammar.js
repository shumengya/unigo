// 生成 syntaxes/garnet.tmLanguage.json。
// 修改语法后运行: node scripts/build-grammar.js

const fs = require('fs');
const path = require('path');

// Garnet 的名字可以是中文等任意字母,所以用 Unicode 属性而不是 \w。
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

const typeCapture = { name: 'meta.type.garnet', patterns: [{ include: '#type-expression' }] };

// 名字 :类型
const annotated = (nameScope) => ({
  match: `${B}(${ID})\\s*(:)\\s*(${TYPE})`,
  captures: {
    1: { name: nameScope },
    2: { name: 'punctuation.separator.type.garnet' },
    3: typeCapture,
  },
});

const block = (keyword, storage, inner) => ({
  begin: `${B}(${keyword})${E}\\s+(${ID})\\s*(\\{)`,
  beginCaptures: {
    1: { name: storage },
    2: { name: 'entity.name.type.garnet' },
    3: { name: 'punctuation.section.braces.begin.garnet' },
  },
  end: '\\}',
  endCaptures: { 0: { name: 'punctuation.section.braces.end.garnet' } },
  name: `meta.${keyword}.garnet`,
  patterns: [{ include: '#comments' }, ...inner, { include: '#punctuation' }],
});

const grammar = {
  $schema: 'https://raw.githubusercontent.com/martinring/tmlanguage/master/tmlanguage.json',
  name: 'Garnet',
  scopeName: 'source.garnet',
  fileTypes: ['gn'],
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
          beginCaptures: { 0: { name: 'punctuation.definition.comment.garnet' } },
          end: '$',
          name: 'comment.line.double-slash.garnet',
          patterns: [
            { match: `${B}(?:TODO|FIXME|XXX|NOTE|HACK)${E}`, name: 'keyword.codetag.notation.garnet' },
          ],
        },
      ],
    },

    strings: {
      patterns: [
        {
          begin: '"',
          beginCaptures: { 0: { name: 'punctuation.definition.string.begin.garnet' } },
          end: '"|$',
          endCaptures: { 0: { name: 'punctuation.definition.string.end.garnet' } },
          name: 'string.quoted.double.garnet',
          patterns: [
            { match: '\\\\[ntr\\\\"]', name: 'constant.character.escape.garnet' },
            { match: '\\\\.', name: 'invalid.illegal.unknown-escape.garnet' },
          ],
        },
        {
          begin: '`',
          beginCaptures: { 0: { name: 'punctuation.definition.string.begin.garnet' } },
          end: '`',
          endCaptures: { 0: { name: 'punctuation.definition.string.end.garnet' } },
          name: 'string.quoted.other.raw.garnet',
        },
        {
          // Garnet 不支持字符字面量。
          match: "'[^'\\n]*'?",
          name: 'invalid.illegal.character-literal.garnet',
        },
      ],
    },

    numbers: {
      patterns: [
        {
          match: `${B}[0-9]+(?:\\.[0-9]+(?:[eE][+-]?[0-9]+)?|[eE][+-]?[0-9]+)${E}`,
          name: 'constant.numeric.float.garnet',
        },
        { match: `${B}[0-9]+${E}`, name: 'constant.numeric.integer.garnet' },
      ],
    },

    'package-clause': {
      match: `${B}(package)${E}\\s+(${ID})`,
      captures: {
        1: { name: 'keyword.other.package.garnet' },
        2: { name: 'entity.name.namespace.garnet' },
      },
    },

    'import-declaration': {
      match: `${B}(import)${E}\\s+(${ID})\\s+(from)${E}\\s*("[^"\\n]*")`,
      captures: {
        1: { name: 'keyword.other.import.garnet' },
        2: { name: 'entity.name.namespace.garnet' },
        3: { name: 'keyword.other.from.garnet' },
        4: { name: 'string.quoted.double.garnet' },
      },
    },

    'struct-declaration': block('struct', 'storage.type.struct.garnet', [
      annotated('variable.other.property.garnet'),
    ]),

    'interface-declaration': block('interface', 'storage.type.interface.garnet', [
      { include: '#function-declaration' },
    ]),

    'enum-declaration': block('enum', 'storage.type.enum.garnet', [
      { match: `${B}${ID}${E}`, name: 'variable.other.enummember.garnet' },
    ]),

    // func 名字(参数) :返回类型
    'function-declaration': {
      begin: `${B}(func)${E}\\s+(${ID})\\s*(\\()`,
      beginCaptures: {
        1: { name: 'storage.type.function.garnet' },
        2: { name: 'entity.name.function.garnet' },
        3: { name: 'punctuation.definition.parameters.begin.garnet' },
      },
      end: `(\\))(?:\\s*(:)\\s*(${TYPE}))?`,
      endCaptures: {
        1: { name: 'punctuation.definition.parameters.end.garnet' },
        2: { name: 'punctuation.separator.type.garnet' },
        3: typeCapture,
      },
      name: 'meta.function.garnet',
      patterns: [
        { include: '#comments' },
        annotated('variable.parameter.garnet'),
        { match: ',', name: 'punctuation.separator.comma.garnet' },
      ],
    },

    'newtype-declaration': {
      match: `${B}(newtype|alias)${E}\\s+(${ID})\\s*(=)\\s*(${TYPE})`,
      captures: {
        1: { name: 'storage.type.garnet' },
        2: { name: 'entity.name.type.garnet' },
        3: { name: 'keyword.operator.assignment.garnet' },
        4: typeCapture,
      },
    },

    // var 名字 :类型 / const 名字 :类型 / var (a :T, b :U)
    'var-declaration': {
      patterns: [
        {
          match: `${B}(var|const)${E}\\s+(${ID})\\s*(:)\\s*(${TYPE})`,
          captures: {
            1: { name: 'storage.type.garnet' },
            2: { name: 'variable.other.declaration.garnet' },
            3: { name: 'punctuation.separator.type.garnet' },
            4: typeCapture,
          },
        },
        {
          begin: `${B}(var)${E}\\s*(\\()`,
          beginCaptures: {
            1: { name: 'storage.type.garnet' },
            2: { name: 'punctuation.section.parens.begin.garnet' },
          },
          end: '\\)',
          endCaptures: { 0: { name: 'punctuation.section.parens.end.garnet' } },
          name: 'meta.var.multiple.garnet',
          patterns: [
            annotated('variable.other.declaration.garnet'),
            { match: ',', name: 'punctuation.separator.comma.garnet' },
          ],
        },
      ],
    },

    'else-return': {
      match: `${B}(else)${E}\\s+(return)${E}`,
      captures: {
        1: { name: 'keyword.control.garnet' },
        2: { name: 'keyword.control.flow.garnet' },
      },
    },

    keywords: {
      patterns: [
        { match: kw(['if', 'elif', 'else', 'switch', 'case', 'default']), name: 'keyword.control.garnet' },
        { match: kw(['while', 'for']), name: 'keyword.control.loop.garnet' },
        { match: kw(['return']), name: 'keyword.control.flow.garnet' },
        { match: kw(['package', 'import', 'from']), name: 'keyword.other.import.garnet' },
        { match: kw(['var', 'const']), name: 'storage.type.garnet' },
        { match: kw(['func']), name: 'storage.type.function.garnet' },
        { match: kw(['enum', 'struct', 'interface', 'newtype', 'alias', 'map']), name: 'storage.type.garnet' },
        { match: kw(['true', 'false']), name: 'constant.language.boolean.garnet' },
        { match: kw(['nil']), name: 'constant.language.null.garnet' },
        { match: kw(FORBIDDEN), name: 'invalid.illegal.forbidden.garnet' },
      ],
    },

    // case Status.ready {
    // 不用变长后顾，VS Code 的 Oniguruma 不接受 (?<=case\s+)。
    'case-clause': {
      match: `${B}(case)${E}\\s+(${ID})\\s*(\\.)\\s*(${ID})`,
      captures: {
        1: { name: 'keyword.control.garnet' },
        2: { name: 'entity.name.type.enum.garnet' },
        3: { name: 'punctuation.accessor.garnet' },
        4: { name: 'variable.other.enummember.garnet' },
      },
    },

    'annotated-name': annotated('variable.other.garnet'),

    // Student{...} 或 StudentParser{ func ... }
    'composite-literal': {
      match: `${B}(?!(?:${KEYWORDS.join('|')})${E})(${ID})(?=\\{)`,
      captures: { 1: { name: 'entity.name.type.garnet' } },
    },

    // Student{name = "树萌芽"} 里的字段名。
    // TextMate 按行匹配，代码块的 { 后换行，块里的赋值不会被误判成字段。
    'field-initializer': {
      match: `([{,])\\s*(${ID})\\s*(?==(?!=))`,
      captures: {
        1: { name: 'punctuation.separator.garnet' },
        2: { name: 'variable.other.property.garnet' },
      },
    },

    calls: {
      patterns: [
        {
          match: `${B}(${BUILTIN_FUNCS.join('|')})\\s*(?=\\()`,
          captures: { 1: { name: 'support.function.builtin.garnet' } },
        },
        {
          // 类型转换，如 string(x)、json.RawMessage(x) 中的内置类型
          match: `${B}(${BUILTIN_TYPES.join('|')})\\s*(?=\\()`,
          captures: { 1: { name: 'support.type.garnet' } },
        },
        {
          // fmt.Println(...) 这类成员调用。不用后顾，避免变长断言。
          match: `(\\.)\\s*(${ID})\\s*(?=\\()`,
          captures: {
            1: { name: 'punctuation.accessor.garnet' },
            2: { name: 'entity.name.function.member.garnet' },
          },
        },
        {
          match: `${B}(?!(?:${KEYWORDS.join('|')})${E})(${ID})\\s*(?=\\()`,
          captures: { 1: { name: 'entity.name.function.garnet' } },
        },
      ],
    },

    'member-access': {
      match: `(\\.)\\s*(${ID})`,
      captures: {
        1: { name: 'punctuation.accessor.garnet' },
        2: { name: 'variable.other.property.garnet' },
      },
    },

    'type-expression': {
      patterns: [
        { match: kw(['map']), name: 'storage.type.map.garnet' },
        { match: kw(BUILTIN_TYPES), name: 'support.type.garnet' },
        {
          match: `(${ID})(\\.)(${ID})`,
          captures: {
            1: { name: 'entity.name.namespace.garnet' },
            2: { name: 'punctuation.accessor.garnet' },
            3: { name: 'entity.name.type.garnet' },
          },
        },
        { match: `${B}${ID}${E}`, name: 'entity.name.type.garnet' },
        { match: '\\*', name: 'keyword.operator.pointer.garnet' },
        { match: '\\[|\\]', name: 'punctuation.section.brackets.garnet' },
        { match: ',', name: 'punctuation.separator.comma.garnet' },
      ],
    },

    operators: {
      patterns: [
        { match: ':=', name: 'invalid.illegal.short-assign.garnet' },
        { match: '/\\*|\\*/', name: 'invalid.illegal.block-comment.garnet' },
        { match: '\\|\\||&&|!(?!=)', name: 'keyword.operator.logical.garnet' },
        { match: '==|!=|<=|>=|<|>', name: 'keyword.operator.comparison.garnet' },
        { match: '[+\\-*/%]', name: 'keyword.operator.arithmetic.garnet' },
        { match: '&', name: 'keyword.operator.address.garnet' },
        { match: '=', name: 'keyword.operator.assignment.garnet' },
      ],
    },

    punctuation: {
      patterns: [
        { match: '\\(', name: 'punctuation.section.parens.begin.garnet' },
        { match: '\\)', name: 'punctuation.section.parens.end.garnet' },
        { match: '\\{', name: 'punctuation.section.braces.begin.garnet' },
        { match: '\\}', name: 'punctuation.section.braces.end.garnet' },
        { match: '\\[', name: 'punctuation.section.brackets.begin.garnet' },
        { match: '\\]', name: 'punctuation.section.brackets.end.garnet' },
        { match: ',', name: 'punctuation.separator.comma.garnet' },
        { match: ';', name: 'punctuation.terminator.statement.garnet' },
        { match: ':', name: 'punctuation.separator.colon.garnet' },
        { match: '\\.', name: 'punctuation.accessor.garnet' },
      ],
    },

    identifiers: {
      match: `${B}${ID}${E}`,
      name: 'variable.other.garnet',
    },
  },
};

const out = path.join(__dirname, '..', 'syntaxes', 'garnet.tmLanguage.json');
fs.writeFileSync(out, JSON.stringify(grammar, null, 2) + '\n');
console.log('wrote', path.relative(process.cwd(), out));
