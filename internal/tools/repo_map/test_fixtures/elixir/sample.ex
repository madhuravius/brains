defmodule Greeter do
  @moduledoc """
  Greeter module docs.
  """

  # Test case: Standard function with a default parameter
  @doc "Says hello."
  def hello(name \\ "world") do
    "Hello " <> name
  end

  # Test case: Standard private function
  defp private_thing(x), do: x

  # Test case: Standard macro
  @doc "Debug macro."
  defmacro debug(expr) do
    quote do: IO.inspect(unquote(expr))
  end

  # --- New Additions for Coverage ---

  # Test case: For the 'defmacrop' symbol type
  @doc "A private macro."
  defmacrop private_macro(val) do
    val
  end

  # Test case: For the 'defprotocol' symbol type
  @doc "A simple protocol."
  defprotocol Parser do
    @doc "Parses the data."
    def parse(data)
  end

  # Test case: For the 'defimpl' symbol type.
  # Also tests qualified names in the definition.
  defimpl Parser, for: Greeter do
    def parse(_data), do: :ok
  end

  # Test case: Function with no arguments to test param parsing
  def no_args do
    :no_args_here
  end

  # Test case: Function defined without parentheses
  def no_parens, do: :no_parens_either

  # Test case: For `@doc false` handling in `elixirLeadingDocAttribute`
  @doc false
  def undocumented_fun do
    "should not have a doc"
  end

  # Test case: To test fallback to `#` comments (`leadingDocComments`)
  def commented_fun do
    "should use the hash comments above as docs"
  end

  # Test case: A doc attribute separated by a blank line should NOT be associated
  @doc "This doc should be ignored."

  def fun_with_separated_doc do
    "no doc here"
  end
end
