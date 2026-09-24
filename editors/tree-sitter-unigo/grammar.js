/**
 * @file UniGo grammar for tree-sitter
 * @license MIT
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

const PREC = {
  or: 1,
  and: 2,
  equality: 3,
  compare: 4,
  additive: 5,
  multiplicative: 6,
  unary: 7,
  postfix: 8,
};

const commaSep = (rule) => optional(seq(rule, repeat(seq(',', rule)), optional(',')));
const commaSep1 = (rule) => seq(rule, repeat(seq(',', rule)), optional(','));

module.exports = grammar({
  name: 'unigo',

  extras: $ => [/\s/, $.comment],

  word: $ => $.identifier,

  supertypes: $ => [$._declaration, $._statement, $._expression, $._type],

  conflicts: $ => [
    [$._expression, $.qualified_type],
  ],

  rules: {
    source_file: $ => seq(
      optional($.package_clause),
      repeat($.import_declaration),
      repeat($._declaration),
    ),

    comment: _ => token(seq('//', /[^\r\n]*/)),

    package_clause: $ => seq('package', field('name', $.identifier), ';'),

    import_declaration: $ => seq(
      'import',
      field('name', $.identifier),
      'from',
      field('path', $.string),
      ';',
    ),

    _declaration: $ => choice(
      $.const_declaration,
      $.var_declaration,
      $.function_declaration,
      $.enum_declaration,
      $.struct_declaration,
      $.interface_declaration,
      $.newtype_declaration,
      $.alias_declaration,
    ),

    // ---------- declarations ----------

    const_declaration: $ => seq(
      'const',
      field('name', $.identifier),
      $.type_annotation,
      '=',
      field('value', $._expression),
      ';',
    ),

    var_declaration: $ => seq(
      'var',
      choice(
        $.var_spec,
        $.var_spec_list,
      ),
      '=',
      field('value', $._expression),
      optional($.else_return),
      ';',
    ),

    var_spec_list: $ => seq('(', commaSep1($.var_spec), ')'),

    var_spec: $ => seq(field('name', $.identifier), $.type_annotation),

    type_annotation: $ => seq(':', field('type', $._type)),

    else_return: $ => seq('else', 'return', optional(field('value', $._expression))),

    function_declaration: $ => seq(
      'func',
      field('name', $.identifier),
      field('parameters', $.parameter_list),
      optional(field('result', $.type_annotation)),
      field('body', $.block),
    ),

    parameter_list: $ => seq('(', commaSep($.parameter), ')'),

    parameter: $ => seq(field('name', $.identifier), $.type_annotation),

    enum_declaration: $ => seq(
      'enum',
      field('name', alias($.identifier, $.type_identifier)),
      field('body', $.enum_body),
    ),

    enum_body: $ => seq('{', repeat(seq($.enum_member, ';')), '}'),

    enum_member: $ => $.identifier,

    struct_declaration: $ => seq(
      'struct',
      field('name', alias($.identifier, $.type_identifier)),
      field('body', $.field_list),
    ),

    field_list: $ => seq('{', repeat($.field_declaration), '}'),

    field_declaration: $ => seq(field('name', $.identifier), $.type_annotation, ';'),

    interface_declaration: $ => seq(
      'interface',
      field('name', alias($.identifier, $.type_identifier)),
      field('body', $.method_list),
    ),

    method_list: $ => seq('{', repeat(seq($.method_signature, ';')), '}'),

    method_signature: $ => seq(
      'func',
      field('name', $.identifier),
      field('parameters', $.parameter_list),
      optional(field('result', $.type_annotation)),
    ),

    newtype_declaration: $ => seq(
      'newtype',
      field('name', alias($.identifier, $.type_identifier)),
      '=',
      field('type', $._type),
      ';',
    ),

    alias_declaration: $ => seq(
      'alias',
      field('name', alias($.identifier, $.type_identifier)),
      '=',
      field('type', $._type),
      ';',
    ),

    // ---------- types ----------

    _type: $ => choice(
      alias($.identifier, $.type_identifier),
      $.qualified_type,
      $.pointer_type,
      $.map_type,
      $.slice_type,
    ),

    qualified_type: $ => seq(
      field('package', $.identifier),
      '.',
      field('name', $.identifier),
    ),

    pointer_type: $ => seq('*', choice(alias($.identifier, $.type_identifier), $.qualified_type)),

    map_type: $ => seq('map', '[', field('key', $._type), ',', field('value', $._type), ']'),

    slice_type: $ => seq(field('element', $._type), '[', ']'),

    // ---------- statements ----------

    block: $ => seq('{', repeat($._statement), '}'),

    _statement: $ => choice(
      $.const_declaration,
      $.var_declaration,
      $.if_statement,
      $.while_statement,
      $.for_statement,
      $.switch_statement,
      $.return_statement,
      $.assignment_statement,
      $.expression_statement,
    ),

    condition: $ => seq('(', $._expression, ')'),

    if_statement: $ => seq(
      'if',
      field('condition', $.condition),
      field('consequence', $.block),
      repeat($.elif_clause),
      optional($.else_clause),
    ),

    elif_clause: $ => seq('elif', field('condition', $.condition), field('consequence', $.block)),

    else_clause: $ => seq('else', field('body', $.block)),

    while_statement: $ => seq('while', field('condition', $.condition), field('body', $.block)),

    for_statement: $ => seq('for', field('condition', $.condition), field('body', $.block)),

    switch_statement: $ => seq(
      'switch',
      field('value', $.condition),
      field('body', $.switch_body),
    ),

    switch_body: $ => seq('{', repeat(choice($.case_clause, $.default_clause)), '}'),

    case_clause: $ => seq('case', field("value", $._expression), field('body', $.block)),


    default_clause: $ => seq('default', field('body', $.block)),

    return_statement: $ => seq('return', optional($._expression), ';'),

    assignment_statement: $ => seq(
      field('left', $._expression),
      '=',
      field('right', $._expression),
      ';',
    ),

    expression_statement: $ => seq($._expression, ';'),

    // ---------- expressions ----------

    _expression: $ => choice(
      $.identifier,
      $.int_literal,
      $.float_literal,
      $.string,
      $.true,
      $.false,
      $.nil,
      $.parenthesized_expression,
      $.unary_expression,
      $.binary_expression,
      $.call_expression,
      $.selector_expression,
      $.index_expression,
      $.slice_literal,
      $.map_literal,
      $.composite_literal,
    ),

    parenthesized_expression: $ => seq('(', $._expression, ')'),

    unary_expression: $ => prec(PREC.unary, seq(
      field('operator', choice('&', '!', '-')),
      field('operand', $._expression),
    )),

    binary_expression: $ => {
      const table = [
        [PREC.or, '||'],
        [PREC.and, '&&'],
        [PREC.equality, choice('==', '!=')],
        [PREC.compare, choice('<', '>', '<=', '>=')],
        [PREC.additive, choice('+', '-')],
        [PREC.multiplicative, choice('*', '/', '%')],
      ];
      return choice(...table.map(([p, op]) => prec.left(p, seq(
        field('left', $._expression),
        // @ts-ignore
        field('operator', op),
        field('right', $._expression),
      ))));
    },

    call_expression: $ => prec(PREC.postfix, seq(
      field('function', $._expression),
      field('arguments', $.argument_list),
    )),

    argument_list: $ => seq('(', commaSep($._expression), ')'),

    selector_expression: $ => prec(PREC.postfix, seq(
      field('operand', $._expression),
      '.',
      field('field', $.identifier),
    )),

    index_expression: $ => prec(PREC.postfix, seq(
      field('operand', $._expression),
      '[',
      field('index', $._expression),
      ']',
    )),

    slice_literal: $ => seq('[', commaSep($._expression), ']'),

    map_literal: $ => seq('{', commaSep($.map_entry), '}'),

    map_entry: $ => seq(field('key', $._expression), ':', field('value', $._expression)),

    composite_literal: $ => prec(PREC.postfix, seq(
      field('type', choice($.identifier, $.qualified_type)),
      field('body', choice($.struct_body, $.implementation_body)),
    )),

    struct_body: $ => seq('{', commaSep($.field_initializer), '}'),

    field_initializer: $ => seq(
      field('name', $.identifier),
      '=',
      field('value', $._expression),
    ),

    implementation_body: $ => seq('{', repeat1(seq($.method_implementation, ';')), '}'),

    method_implementation: $ => seq(
      'func',
      field('name', $.identifier),
      field('parameters', $.parameter_list),
      optional(field('result', $.type_annotation)),
      field('body', $.block),
    ),

    // ---------- literals & names ----------

    string: $ => choice($._interpreted_string, $.raw_string),

    _interpreted_string: $ => seq(
      '"',
      repeat(choice(
        token.immediate(prec(1, /[^"\\\r\n]+/)),
        $.escape_sequence,
      )),
      token.immediate('"'),
    ),

    escape_sequence: _ => token.immediate(/\\./),

    raw_string: _ => token(seq('`', /[^`]*/, '`')),

    int_literal: _ => token(/[0-9]+/),

    float_literal: _ => token(choice(
      /[0-9]+\.[0-9]+([eE][+-]?[0-9]+)?/,
      /[0-9]+[eE][+-]?[0-9]+/,
    )),

    true: _ => 'true',
    false: _ => 'false',
    nil: _ => 'nil',

    identifier: _ => /[_\p{L}][_\p{L}\p{Nd}]*/,

  },
});
