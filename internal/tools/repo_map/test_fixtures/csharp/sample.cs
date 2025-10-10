namespace Example {
    public struct Point {
        public int X;
        public int Y;
    }

    public interface IRepo {
        void Save(object o);
    }

    public class Person {
        public void Speak() { System.Console.WriteLine("hi"); }
    }
}

