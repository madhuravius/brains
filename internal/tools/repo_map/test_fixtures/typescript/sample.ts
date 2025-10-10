export interface IThing {
  id: number;
  name: string;
}

export class MyClass {
  do() { return "done"; }
}

export function doThing(x: number): number {
  return x * 2;
}

