const VERSION: &str = "1.0";

struct Person {
    name: String,
}

impl Person {
    fn greet(&self) -> String {
        format!("Hello, {}", self.name)
    }
}

fn add(a: i32, b: i32) -> i32 {
    a + b
}

trait Speak {
    fn speak(&self);
}

