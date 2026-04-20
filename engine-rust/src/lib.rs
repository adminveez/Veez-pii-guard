// veez-pii-engine — optional Rust backend for veez-pii-guard.
//
// FFI surface (per ADR-006):
//
//   pub extern "C" fn veez_pii_scan(
//       input_ptr: *const u8,
//       input_len: usize,
//       out_ptr: *mut *mut u8,
//       out_len: *mut usize,
//   ) -> i32;
//
//   pub extern "C" fn veez_pii_free(ptr: *mut u8, len: usize);
//
// Wire format: input is UTF-8 bytes; output is JSON-encoded
// `Vec<Detection>` allocated on the Rust heap. The caller MUST call
// `veez_pii_free` with the returned ptr+len once done. We allocate via
// `Vec::into_raw_parts`-style manual layout so Go's cgo can consume.
//
// Feature parity with the Go engine is intentionally narrow in v0.2:
// only the high-volume regex passes (email, phone, IBAN, IP) are routed
// here. Plugin invocations and contextual name detection remain in Go.

use once_cell::sync::Lazy;
use regex::Regex;
use serde::Serialize;
use std::slice;

#[derive(Serialize)]
struct Detection {
    r#type: &'static str,
    text: String,
    start: usize,
    end: usize,
    confidence: f32,
    source: &'static str,
}

static EMAIL: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"\b[\w.+-]+@[\w-]+\.[\w.-]+\b").unwrap()
});

static PHONE_FR: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"(?:\+33|0)[ .-]?[1-9](?:[ .-]?\d{2}){4}").unwrap()
});

static IPV4: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"\b(?:\d{1,3}\.){3}\d{1,3}\b").unwrap()
});

fn scan(input: &str) -> Vec<Detection> {
    let mut out = Vec::new();
    for m in EMAIL.find_iter(input) {
        out.push(Detection {
            r#type: "EMAIL",
            text: m.as_str().to_string(),
            start: m.start(),
            end: m.end(),
            confidence: 0.95,
            source: "rust:email",
        });
    }
    for m in PHONE_FR.find_iter(input) {
        out.push(Detection {
            r#type: "PHONE",
            text: m.as_str().to_string(),
            start: m.start(),
            end: m.end(),
            confidence: 0.85,
            source: "rust:phone_fr",
        });
    }
    for m in IPV4.find_iter(input) {
        out.push(Detection {
            r#type: "IP",
            text: m.as_str().to_string(),
            start: m.start(),
            end: m.end(),
            confidence: 0.80,
            source: "rust:ipv4",
        });
    }
    out
}

/// Scan UTF-8 bytes for PII. On success, writes a JSON-encoded
/// `Vec<Detection>` into a freshly-allocated buffer and returns 0.
/// The caller MUST free the buffer via `veez_pii_free`.
///
/// # Safety
/// `input_ptr` must point to `input_len` valid UTF-8 bytes.
/// `out_ptr` and `out_len` must be valid, non-null pointers.
#[no_mangle]
pub unsafe extern "C" fn veez_pii_scan(
    input_ptr: *const u8,
    input_len: usize,
    out_ptr: *mut *mut u8,
    out_len: *mut usize,
) -> i32 {
    if input_ptr.is_null() || out_ptr.is_null() || out_len.is_null() {
        return -1;
    }
    let bytes = slice::from_raw_parts(input_ptr, input_len);
    let input = match std::str::from_utf8(bytes) {
        Ok(s) => s,
        Err(_) => return -2,
    };
    let detections = scan(input);
    let json = match serde_json::to_vec(&detections) {
        Ok(v) => v,
        Err(_) => return -3,
    };
    let len = json.len();
    let mut boxed = json.into_boxed_slice();
    let ptr = boxed.as_mut_ptr();
    std::mem::forget(boxed);
    *out_ptr = ptr;
    *out_len = len;
    0
}

/// Free a buffer previously allocated by `veez_pii_scan`.
///
/// # Safety
/// `ptr` must come from `veez_pii_scan` and `len` must match.
#[no_mangle]
pub unsafe extern "C" fn veez_pii_free(ptr: *mut u8, len: usize) {
    if ptr.is_null() {
        return;
    }
    let _ = Box::from_raw(slice::from_raw_parts_mut(ptr, len));
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn finds_email() {
        let d = scan("contact alice@example.com today");
        assert_eq!(d.len(), 1);
        assert_eq!(d[0].r#type, "EMAIL");
    }

    #[test]
    fn finds_phone_and_ip() {
        let d = scan("call +33 6 12 34 56 78 from 10.0.0.1");
        assert!(d.iter().any(|x| x.r#type == "PHONE"));
        assert!(d.iter().any(|x| x.r#type == "IP"));
    }
}
