// Code generated by schema tool; DO NOT EDIT.

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

#![allow(dead_code)]
#![allow(unused_imports)]

use crate::*;
use crate::coreroot::*;

#[derive(Clone)]
pub struct ImmutableFindContractResults {
    pub proxy: Proxy,
}

impl ImmutableFindContractResults {
    // whether or not the contract exists.
    pub fn contract_found(&self) -> ScImmutableBool {
        ScImmutableBool::new(self.proxy.root(RESULT_CONTRACT_FOUND))
    }

    // encoded contract record (if exists)
    pub fn contract_rec_data(&self) -> ScImmutableBytes {
        ScImmutableBytes::new(self.proxy.root(RESULT_CONTRACT_REC_DATA))
    }
}

#[derive(Clone)]
pub struct MutableFindContractResults {
    pub proxy: Proxy,
}

impl MutableFindContractResults {
    pub fn new() -> MutableFindContractResults {
        MutableFindContractResults {
            proxy: results_proxy(),
        }
    }

    // whether or not the contract exists.
    pub fn contract_found(&self) -> ScMutableBool {
        ScMutableBool::new(self.proxy.root(RESULT_CONTRACT_FOUND))
    }

    // encoded contract record (if exists)
    pub fn contract_rec_data(&self) -> ScMutableBytes {
        ScMutableBytes::new(self.proxy.root(RESULT_CONTRACT_REC_DATA))
    }
}

#[derive(Clone)]
pub struct MapHnameToImmutableBytes {
    pub(crate) proxy: Proxy,
}

impl MapHnameToImmutableBytes {
    pub fn get_bytes(&self, key: ScHname) -> ScImmutableBytes {
        ScImmutableBytes::new(self.proxy.key(&hname_to_bytes(key)))
    }
}

#[derive(Clone)]
pub struct ImmutableGetContractRecordsResults {
    pub proxy: Proxy,
}

impl ImmutableGetContractRecordsResults {
    // contract records by Hname
    pub fn contract_registry(&self) -> MapHnameToImmutableBytes {
        MapHnameToImmutableBytes { proxy: self.proxy.root(RESULT_CONTRACT_REGISTRY) }
    }
}

#[derive(Clone)]
pub struct MapHnameToMutableBytes {
    pub(crate) proxy: Proxy,
}

impl MapHnameToMutableBytes {
    pub fn clear(&self) {
        self.proxy.clear_map();
    }

    pub fn get_bytes(&self, key: ScHname) -> ScMutableBytes {
        ScMutableBytes::new(self.proxy.key(&hname_to_bytes(key)))
    }
}

#[derive(Clone)]
pub struct MutableGetContractRecordsResults {
    pub proxy: Proxy,
}

impl MutableGetContractRecordsResults {
    pub fn new() -> MutableGetContractRecordsResults {
        MutableGetContractRecordsResults {
            proxy: results_proxy(),
        }
    }

    // contract records by Hname
    pub fn contract_registry(&self) -> MapHnameToMutableBytes {
        MapHnameToMutableBytes { proxy: self.proxy.root(RESULT_CONTRACT_REGISTRY) }
    }
}
