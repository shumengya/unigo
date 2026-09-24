/**
 * ============================================================================
 * TypeScript 语法总览（学习示例）
 * ============================================================================
 *
 * JS 基础（变量、函数、类、async…）见 javascript.js。
 * 这里只写 TypeScript 多出来的「类型语法」。
 *
 * 运行（Node 22+ 会擦掉类型再执行）：
 *   node typescript.ts
 *
 * 记住：类型只在编译期存在，运行时全部擦掉。
 *
 * 同目录还有 javascript.js。没有 tsconfig 时，编辑器会把它们当成「一个全局脚本」，
 * 两边的 const / function / interface 会互相撞名报错。本目录的 tsconfig.json
 * 只把本文件纳入检查。
 */

// ============================================================================
// 1. 基本类型
// ============================================================================

const tStr: string = "hello";
const tNum: number = 42; // 含整数和小数；没有 int
const tBool: boolean = true;
const tNull: null = null;
const tUndef: undefined = undefined;
const tBig: bigint = 10n;
const tSym: symbol = Symbol("id");

// any：关掉类型检查，能写但尽量不用
let tAny: any = 1;
tAny = "ok";

// unknown：类型安全的 any。用之前必须收窄
let tUnk: unknown = JSON.parse('{"n":1}');

// void：函数「没有有意义的返回值」
function logMsg(msg: string): void {
  console.log(msg);
}

// never：不可能返回（抛错或死循环）
function fail(msg: string): never {
  throw new Error(msg);
}

// ============================================================================
// 2. 数组、元组
// ============================================================================

const nums: number[] = [1, 2, 3];
const nums2: Array<number> = [4, 5]; // 同一意思，习惯写 number[]

const pair: [string, number] = ["Ada", 36]; // 元组：长度和每项类型都固定
const point: [x: number, y: number] = [1, 2]; // 可给元素起名，纯文档
pair[0].toUpperCase();
pair[1].toFixed(0);

const ro: readonly number[] = [1, 2, 3]; // 只读，不能 push

// ============================================================================
// 3. 对象类型、可选、只读、索引
// ============================================================================

type User = {
  readonly id: number; // 创建后不能改
  name: string;
  age?: number; // 可选，可以没有，值也可能是 undefined
  addr?: { city: string };
};

const u1: User = { id: 1, name: "Ada" };
const u2: User = { id: 2, name: "Bob", age: 20 };

// 索引签名：键的形状
type Dict = { [key: string]: number };
const scores: Dict = { math: 90, en: 88 };

// ============================================================================
// 4. type vs interface
// ============================================================================

// 两者都能描述对象。习惯：对象形状用 interface，联合/元组/映射用 type。
interface Animal {
  name: string;
}

interface Dog extends Animal {
  bark(): string;
}

type Cat = Animal & { meow(): string }; // 交叉类型 ≈ 合并字段

interface Box<T> {
  value: T;
}

const dog: Dog = { name: "Fido", bark: () => "woof" };
const box: Box<number> = { value: 1 };

// interface 同名会合并（声明合并）；type 不行。一般别依赖合并。

// ============================================================================
// 5. 联合、字面量、判别联合
// ============================================================================

type ID = string | number; // 联合：其中之一
function printId(id: ID): string {
  if (typeof id === "string") return id.toUpperCase(); // 收窄后才能当 string 用
  return id.toFixed(0);
}

type Align = "left" | "right" | "center"; // 字符串字面量联合（比 enum 更常用）
const align: Align = "left";

type Status =
  | { kind: "ok"; data: string }
  | { kind: "err"; error: string }; // 判别联合：用共同字段区分

function unwrap(s: Status): string {
  switch (s.kind) {
    case "ok":
      return s.data;
    case "err":
      return s.error;
  }
}

// ============================================================================
// 6. 函数类型
// ============================================================================

type Binary = (a: number, b: number) => number;
const add: Binary = (a, b) => a + b;

function greet(name: string, title?: string): string {
  return title ? `${title} ${name}` : name;
}

function greet2(name: string, title: string = "Ms"): string {
  return `${title} ${name}`;
}

function sumAll(...nums: number[]): number {
  return nums.reduce((a, b) => a + b, 0);
}

// 函数重载：给调用方看的是上面几条，真正实现是最后一条
function len(x: string): number;
function len(x: unknown[]): number;
function len(x: string | unknown[]): number {
  return x.length;
}

async function load(id: number): Promise<User> {
  return { id, name: "Ada" };
}

// ============================================================================
// 7. 类（TS 多出来的部分）
// ============================================================================

class Counter {
  n: number = 0;
  readonly created: number = Date.now();
  private secret = 1; // 仅类型检查；编译后仍是普通字段（不是 # 真私有）
  protected kind = "counter";

  inc(): this {
    this.n++;
    return this; // this 类型：链式调用保持子类类型
  }
}

class NamedCounter extends Counter {
  name: string;
  constructor(name: string) {
    super();
    this.name = name;
  }
}

// 语法糖（需要 tsc 转换，node 直接擦类型时不支持）：
//   constructor(public name: string) {}  声明字段 + this.name = name
//   abstract class Shape { abstract area(): number }

interface Printable {
  print(): string;
}

class Label implements Printable {
  private readonly title: string;
  constructor(title: string) {
    this.title = title;
  }
  print(): string {
    return this.title;
  }
}

// ============================================================================
// 8. 泛型
// ============================================================================

function identity<T>(x: T): T {
  return x;
}

function first<T>(xs: T[]): T | undefined {
  return xs[0];
}

function pickProp<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}

class Stack<T> {
  private items: T[] = [];
  push(x: T): void {
    this.items.push(x);
  }
  pop(): T | undefined {
    return this.items.pop();
  }
}

// 约束：T 必须有 length
function longer<T extends { length: number }>(a: T, b: T): T {
  return a.length >= b.length ? a : b;
}

const idNum = identity(1); // 推断 T = number
const idStr = identity<string>("a"); // 显式传入

// ============================================================================
// 9. 类型收窄
// ============================================================================

function pad(x: string | number): string {
  if (typeof x === "number") return x.toFixed(0); // typeof
  return x.trim();
}

function isUser(v: unknown): v is User {
  // 类型谓词：return true 时，调用方把 v 当成 User
  return typeof v === "object" && v !== null && "id" in v && "name" in v;
}

function area(x: { kind: "circle"; r: number } | { kind: "rect"; w: number; h: number }): number {
  if (x.kind === "circle") return Math.PI * x.r ** 2; // 判别字段
  return x.w * x.h;
}

function maybeName(u: User): string {
  return u.addr?.city ?? "unknown"; // 可选链 + 空值合并（JS 语法，TS 会跟着收窄）
}

// ============================================================================
// 10. typeof / keyof / 索引访问 / 映射
// ============================================================================

const defaults = { host: "127.0.0.1", port: 8080 };
type Defaults = typeof defaults; // { host: string; port: number }
type DefaultKey = keyof Defaults; // "host" | "port"

type UserName = User["name"]; // 索引访问 → string
type UserIdOrName = User["id" | "name"]; // number | string

// 映射类型：把每个字段变成可选（手写版 Partial）
type MyPartial<T> = { [K in keyof T]?: T[K] };
type UserPatch = MyPartial<User>;

// ============================================================================
// 11. 常用工具类型（标准库自带，直接用）
// ============================================================================

type TPartial = Partial<User>; // 全变成可选
type TRequired = Required<User>; // 全变成必填
type TReadonly = Readonly<User>;
type TPick = Pick<User, "id" | "name">;
type TOmit = Omit<User, "addr">;
type TRecord = Record<"a" | "b", number>; // { a: number; b: number }
type TExclude = Exclude<"a" | "b" | "c", "c">; // "a" | "b"
type TExtract = Extract<"a" | number, string>; // "a"
type TNonNull = NonNullable<string | null | undefined>; // string
type TReturn = ReturnType<typeof add>; // number
type TParams = Parameters<typeof greet>; // [name: string, title?: string]
type TAwait = Awaited<Promise<User>>; // User

const patch: TPick = { id: 1, name: "Ada" };
const rec: TRecord = { a: 1, b: 2 };

// ============================================================================
// 12. as、as const、satisfies、非空断言
// ============================================================================

const el = { n: 1 } as unknown as User; // 双重断言：先 unknown 再 User。容易撒谎，少用
const names = ["a", "b"] as const; // 字面量只读：类型是 readonly ["a", "b"]
type Name = (typeof names)[number]; // "a" | "b"

const cfg = {
  host: "localhost",
  port: 8080,
} satisfies { host: string; port: number };
// satisfies：检查符合形状，但保留更精确的推断（port 仍是 8080 不是 number）

function needUser(u: User | undefined): number {
  return u!.id; // 非空断言：你保证不是 null/undefined，真是空会运行时报错
}

// ============================================================================
// 13. 模块（类型导入）
// ============================================================================

// import type { User } from "./types.ts";  只导入类型，运行时会被擦掉
// import { type User, load } from "./api.ts";  同一句里混用
// export type { User };
// export interface Repo { get(id: number): Promise<User> }

// ============================================================================
// 14. 枚举：能用，但更推荐字面量联合
// ============================================================================

// enum Color { Red, Green }           数字枚举，有反向映射，行为怪
// const enum Color { Red = "red" }    const enum 编译后会内联
// 日常用：type Color = "red" | "green"

type Color = "red" | "green";
const color: Color = "red";

// ============================================================================
// 把关键结果跑一遍
// ============================================================================

void logMsg; // 避免「未使用」告警（视检查器而定）

const stack = new Stack<number>();
stack.push(1);
stack.push(2);

const nc = new NamedCounter("c").inc().inc();
const unk: unknown = { id: 9, name: "Neo" };

const demo = {
  printId: printId("ab"),
  unwrap: unwrap({ kind: "ok", data: "yes" }),
  add: add(2, 3),
  greet: greet("Ada", "Dr"),
  greet2: greet2("Ada"),
  sumAll: sumAll(1, 2, 3),
  len: len("hey"),
  identity: idNum,
  first: first(nums),
  pick: pickProp(u1, "name"),
  longer: longer("go", "typescript"),
  pad: pad(3),
  isUser: isUser(unk),
  area: area({ kind: "rect", w: 3, h: 4 }),
  city: maybeName(u1),
  stack: stack.pop(),
  counter: nc.n,
  label: new Label("hi").print(),
  align,
  color,
  names: [...names],
  cfg: cfg.host,
  patch,
  rec,
  needUser: needUser(u1),
};

console.log("TypeScript 语法示例运行结果：");
console.log(demo);
