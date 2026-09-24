(package_clause
  "package" @context
  name: (identifier) @name) @item

(import_declaration
  "import" @context
  name: (identifier) @name) @item

(function_declaration
  "func" @context
  name: (identifier) @name
  parameters: (parameter_list) @context) @item

(struct_declaration
  "struct" @context
  name: (type_identifier) @name) @item

(interface_declaration
  "interface" @context
  name: (type_identifier) @name) @item

(enum_declaration
  "enum" @context
  name: (type_identifier) @name) @item

(newtype_declaration
  "newtype" @context
  name: (type_identifier) @name) @item

(alias_declaration
  "alias" @context
  name: (type_identifier) @name) @item

(source_file
  (const_declaration
    "const" @context
    name: (identifier) @name) @item)

(source_file
  (var_declaration
    "var" @context
    (var_spec name: (identifier) @name)) @item)
