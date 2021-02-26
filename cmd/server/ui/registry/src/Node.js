import React, { Component } from "react";
import { Button, Form } from  'semantic-ui-react';

export class AddNodeForm extends Component {
    constructor(props) {
        super(props)
        this.state = { name: "", address: ""}
    }

    handleChange = (e, { name, value }) => {
        this.setState({ [name]: value })
    }

    handleSubmit = () => {
      const { name, address } = this.state
      const url = this.props.service + "/node"
      this.setState({ name: name, address: address })
      fetch(url, { method: "POST", headers: {"Content-Type": "application/json"}, body: JSON.stringify(this.state) })
        .then(resp => { if (resp.ok) { resp.json() } else { console.error(resp) }})
        .catch(err => { console.error(err) })
    }

    render() {
        return <Form onSubmit={this.handleSubmit}>
            <Form.Field>
                <label>Name</label>
                <Form.Input name='name' value={this.state.name} onChange={this.handleChange} placeholder="Node name eg: my-node.srv" />
            </Form.Field>
            <Form.Field>
                <label>Address</label>
                <Form.Input name='address' value={this.state.address} onChange={this.handleChange} placeholder="Node address eg: registry-ui.com.au:8000" />
            </Form.Field>
            <Form.Field>
                <Button type='submit'>Register</Button>
            </Form.Field>
        </Form>
    }
}
