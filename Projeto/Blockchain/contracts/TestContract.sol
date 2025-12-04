// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

contract TestContract {
    uint256 public value;
    address public owner;

    constructor() {
        owner = msg.sender;
        value = 100;
    }

    function setValue(uint256 _newValue) public {
        value = _newValue;
    }
}


