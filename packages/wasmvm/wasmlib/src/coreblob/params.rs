// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the schema definition file instead

#![allow(dead_code)]
#![allow(unused_imports)]

use crate::*;
use crate::coreblob::*;

#[derive(Clone)]
pub struct MapStringToImmutableBytes {
    pub(crate) proxy: Proxy,
}

impl MapStringToImmutableBytes {
    pub fn get_bytes(&self, key: &str) -> ScImmutableBytes {
        ScImmutableBytes::new(self.proxy.key(&string_to_bytes(key)))
    }
}

#[derive(Clone)]
pub struct ImmutableStoreBlobParams {
    pub(crate) proxy: Proxy,
}

impl ImmutableStoreBlobParams {
    pub fn new() -> ImmutableStoreBlobParams {
        ImmutableStoreBlobParams {
            proxy: params_proxy(),
        }
    }

    // named chunks
    pub fn blobs(&self) -> MapStringToImmutableBytes {
        MapStringToImmutableBytes { proxy: self.proxy.clone() }
    }

    // data schema for external tools
    pub fn data_schema(&self) -> ScImmutableBytes {
        ScImmutableBytes::new(self.proxy.root(PARAM_DATA_SCHEMA))
    }

    // smart contract program binary code
    pub fn prog_binary(&self) -> ScImmutableBytes {
        ScImmutableBytes::new(self.proxy.root(PARAM_PROG_BINARY))
    }

    // smart contract program source code
    pub fn sources(&self) -> ScImmutableBytes {
        ScImmutableBytes::new(self.proxy.root(PARAM_SOURCES))
    }

    // VM type that must be used to run progBinary
    pub fn vm_type(&self) -> ScImmutableString {
        ScImmutableString::new(self.proxy.root(PARAM_VM_TYPE))
    }
}

#[derive(Clone)]
pub struct MapStringToMutableBytes {
    pub(crate) proxy: Proxy,
}

impl MapStringToMutableBytes {
    pub fn clear(&self) {
        self.proxy.clear_map();
    }

    pub fn get_bytes(&self, key: &str) -> ScMutableBytes {
        ScMutableBytes::new(self.proxy.key(&string_to_bytes(key)))
    }
}

#[derive(Clone)]
pub struct MutableStoreBlobParams {
    pub(crate) proxy: Proxy,
}

impl MutableStoreBlobParams {
    // named chunks
    pub fn blobs(&self) -> MapStringToMutableBytes {
        MapStringToMutableBytes { proxy: self.proxy.clone() }
    }

    // data schema for external tools
    pub fn data_schema(&self) -> ScMutableBytes {
        ScMutableBytes::new(self.proxy.root(PARAM_DATA_SCHEMA))
    }

    // smart contract program binary code
    pub fn prog_binary(&self) -> ScMutableBytes {
        ScMutableBytes::new(self.proxy.root(PARAM_PROG_BINARY))
    }

    // smart contract program source code
    pub fn sources(&self) -> ScMutableBytes {
        ScMutableBytes::new(self.proxy.root(PARAM_SOURCES))
    }

    // VM type that must be used to run progBinary
    pub fn vm_type(&self) -> ScMutableString {
        ScMutableString::new(self.proxy.root(PARAM_VM_TYPE))
    }
}

#[derive(Clone)]
pub struct ImmutableGetBlobFieldParams {
    pub(crate) proxy: Proxy,
}

impl ImmutableGetBlobFieldParams {
    pub fn new() -> ImmutableGetBlobFieldParams {
        ImmutableGetBlobFieldParams {
            proxy: params_proxy(),
        }
    }

    // chunk name
    pub fn field(&self) -> ScImmutableString {
        ScImmutableString::new(self.proxy.root(PARAM_FIELD))
    }

    // hash of the blob
    pub fn hash(&self) -> ScImmutableHash {
        ScImmutableHash::new(self.proxy.root(PARAM_HASH))
    }
}

#[derive(Clone)]
pub struct MutableGetBlobFieldParams {
    pub(crate) proxy: Proxy,
}

impl MutableGetBlobFieldParams {
    // chunk name
    pub fn field(&self) -> ScMutableString {
        ScMutableString::new(self.proxy.root(PARAM_FIELD))
    }

    // hash of the blob
    pub fn hash(&self) -> ScMutableHash {
        ScMutableHash::new(self.proxy.root(PARAM_HASH))
    }
}

#[derive(Clone)]
pub struct ImmutableGetBlobInfoParams {
    pub(crate) proxy: Proxy,
}

impl ImmutableGetBlobInfoParams {
    pub fn new() -> ImmutableGetBlobInfoParams {
        ImmutableGetBlobInfoParams {
            proxy: params_proxy(),
        }
    }

    // hash of the blob
    pub fn hash(&self) -> ScImmutableHash {
        ScImmutableHash::new(self.proxy.root(PARAM_HASH))
    }
}

#[derive(Clone)]
pub struct MutableGetBlobInfoParams {
    pub(crate) proxy: Proxy,
}

impl MutableGetBlobInfoParams {
    // hash of the blob
    pub fn hash(&self) -> ScMutableHash {
        ScMutableHash::new(self.proxy.root(PARAM_HASH))
    }
}
