/// Person doc
struct Person {
    name: String,
}

/// add doc
fn add(x: i32, y: i32) -> i32 {
    x + y
}

/// Speak trait
trait Speak {
    fn say(&self, msg: &str);
}

