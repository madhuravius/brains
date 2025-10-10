public interface IRepo {}

public struct Point { public int X; public int Y; }

public class Person {
    /// Speak doc
    public void Speak(this string name, ref int count, out string value, params string[] rest) {
        count = 0;
        value = name;
    }
}
