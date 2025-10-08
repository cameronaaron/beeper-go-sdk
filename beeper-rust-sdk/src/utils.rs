/// Utility functions for working with optional values

/// Create an optional string
pub fn optional_string(s: impl Into<String>) -> Option<String> {
    let s = s.into();
    if s.is_empty() {
        None
    } else {
        Some(s)
    }
}

/// Create an optional integer
pub fn optional_i32(i: i32) -> Option<i32> {
    if i == 0 {
        None
    } else {
        Some(i)
    }
}

/// Create an optional usize
pub fn optional_usize(i: usize) -> Option<usize> {
    if i == 0 {
        None
    } else {
        Some(i)
    }
}

/// Create an optional boolean (None if false)
pub fn optional_bool(b: bool) -> Option<bool> {
    if b {
        Some(b)
    } else {
        None
    }
}

/// Join optional vector of strings with a separator
pub fn join_optional_strings(vec: &[String], separator: &str) -> Option<String> {
    if vec.is_empty() {
        None
    } else {
        Some(vec.join(separator))
    }
}

/// Convert query parameter value to string
pub fn query_param_to_string<T: std::fmt::Display>(value: T) -> String {
    value.to_string()
}

/// Convert optional query parameter value to optional string
pub fn optional_query_param<T: std::fmt::Display>(value: Option<T>) -> Option<String> {
    value.map(|v| v.to_string())
}

/// Convert slice to indexed query parameters (for array-style parameters)
pub fn slice_to_indexed_params(key: &str, values: &[String]) -> Vec<(String, String)> {
    values
        .iter()
        .enumerate()
        .map(|(i, value)| (format!("{}[{}]", key, i), value.clone()))
        .collect()
}

/// URL-safe base64 encoding
pub fn base64_encode(data: &[u8]) -> String {
    use base64::Engine;
    base64::engine::general_purpose::URL_SAFE_NO_PAD.encode(data)
}

/// URL-safe base64 decoding
pub fn base64_decode(data: &str) -> Result<Vec<u8>, base64::DecodeError> {
    use base64::Engine;
    base64::engine::general_purpose::URL_SAFE_NO_PAD.decode(data)
}