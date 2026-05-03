"""Unit tests for ndjson_sql_helpers — Bug #20."""
import math
from decimal import Decimal

import pytest

from omni_build.ndjson_sql_helpers import (
    _escape_string,
    pg_literal,
    validate_identifier,
)


class TestEscapeString:
    """Tests for _escape_string covering edge cases."""

    def test_simple_string(self):
        assert _escape_string("hello") == "'hello'"

    def test_single_quote(self):
        assert _escape_string("it's") == "'it''s'"

    def test_double_single_quote(self):
        assert _escape_string("it''s") == "'it''''s'"

    def test_backslash(self):
        result = _escape_string("path\\to\\file")
        assert result.startswith("E'")
        assert result.endswith("'")
        assert "\\\\to\\\\" in result

    def test_backslash_and_quote(self):
        result = _escape_string("it's a \\path")
        assert result.startswith("E'")
        assert "''" in result
        assert "\\\\" in result

    def test_null_bytes_stripped(self):
        result = _escape_string("hello\x00world")
        assert "\x00" not in result
        assert result == "'helloworld'"

    def test_null_byte_with_backslash(self):
        result = _escape_string("\x00back\\slash\x00")
        assert "\x00" not in result
        assert result.startswith("E'")
        assert "\\\\" in result

    def test_unicode_preserved(self):
        result = _escape_string("日本語テスト")
        assert result == "'日本語テスト'"

    def test_unicode_with_quotes(self):
        result = _escape_string("l'été")
        assert result == "'l''été'"

    def test_empty_string(self):
        assert _escape_string("") == "''"

    def test_only_null_bytes(self):
        assert _escape_string("\x00\x00\x00") == "''"

    def test_multiple_backslashes(self):
        result = _escape_string("\\\\\\")
        assert result.startswith("E'")
        # 3 backslashes → 6 escaped
        assert "E'" + "\\\\" * 3 + "'" == result

    def test_newline_preserved(self):
        result = _escape_string("line1\nline2")
        assert "line1\nline2" in result or result == "'line1\nline2'"

    def test_tab_preserved(self):
        result = _escape_string("col1\tcol2")
        assert result == "'col1\tcol2'"


class TestPgLiteral:
    """Tests for pg_literal covering type handling."""

    def test_none(self):
        assert pg_literal(None) == "NULL"

    def test_bool_true(self):
        assert pg_literal(True) == "TRUE"

    def test_bool_false(self):
        assert pg_literal(False) == "FALSE"

    def test_int(self):
        assert pg_literal(42) == "42"

    def test_decimal(self):
        assert pg_literal(Decimal("3.14")) == "3.14"

    def test_float_normal(self):
        assert pg_literal(1.5) == "1.5"

    def test_float_nan(self):
        assert pg_literal(float('nan')) == "'NaN'::float"

    def test_float_inf(self):
        assert pg_literal(float('inf')) == "'Infinity'::float"

    def test_float_neg_inf(self):
        assert pg_literal(float('-inf')) == "'-Infinity'::float"

    def test_dict_json(self):
        result = pg_literal({"key": "value"})
        assert "'key'" not in result  # Should be JSON inside SQL string
        assert "key" in result

    def test_list_json(self):
        result = pg_literal([1, 2, 3])
        assert "[1, 2, 3]" in result

    def test_string(self):
        assert pg_literal("hello") == "'hello'"

    def test_string_with_quote(self):
        assert pg_literal("it's") == "'it''s'"


class TestValidateIdentifier:
    """Tests for validate_identifier."""

    def test_valid_simple(self):
        assert validate_identifier("users") == '"users"'

    def test_valid_underscore(self):
        assert validate_identifier("_private") == '"_private"'

    def test_valid_alphanumeric(self):
        assert validate_identifier("table_123") == '"table_123"'

    def test_invalid_starts_with_number(self):
        with pytest.raises(ValueError):
            validate_identifier("123table")

    def test_invalid_special_chars(self):
        with pytest.raises(ValueError):
            validate_identifier("table; DROP TABLE users")

    def test_invalid_dash(self):
        with pytest.raises(ValueError):
            validate_identifier("my-table")

    def test_invalid_dot(self):
        with pytest.raises(ValueError):
            validate_identifier("schema.table")

    def test_invalid_empty(self):
        with pytest.raises(ValueError):
            validate_identifier("")
