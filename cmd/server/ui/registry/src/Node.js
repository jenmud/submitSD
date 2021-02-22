import React, { Component } from "react";

export class Node extends Component {
    constructor(props) {
        super(props)
        this.state = {
            uid: "UnsetUID",
            name: "UnsetName",
            address: "UnsetAddress",
            meta: {},
        }
    }

    render() {
        return this.state.uid;
    }
}
