CONSTANT = 42

def greet(name):
    """Greet someone."""
    return f"Hello {name}"

class Person:
    """Represents a person."""

    def __init__(self, name):
        self.name = name

    def speak(self):
        """Speak name."""
        return f"My name is {self.name}"
