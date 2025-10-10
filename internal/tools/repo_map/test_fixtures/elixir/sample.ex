defmodule Greeter do
  @moduledoc """
  Greeter module docs.
  """

  @doc "Says hello."
  def hello(name \\ "world") do
    "Hello " <> name
  end

  defp private_thing(x), do: x

  @doc "Debug macro."
  defmacro debug(expr) do
    quote do: IO.inspect(unquote(expr))
  end
end

