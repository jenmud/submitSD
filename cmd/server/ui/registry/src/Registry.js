import React, { Component } from 'react';
import { Button, Checkbox, Icon, Modal, Popup, Table } from 'semantic-ui-react';

import { AddNodeForm } from './Node';


export class RegistryRow extends Component {
    constructor(props) {
        super(props);
        this.state = { value: 'enabled' };
    }

    toggle() {
        if (this.state.value !== 'disabled') {
            fetch(this.props.service + "/node/" + this.props.uid, { method: 'DELETE' })
                .then(resp => {
                    console.log(resp)
                    if (resp.ok) {
                        this.setState({ value: 'disabled' })
                    }
                })
                .catch(err => { console.error(err) })
        }
    }

    handleChange = (event, { value }) => {
        console.debug(event)
        this.toggle()
    }

    render() {
        return (
            <Table.Row id={this.props.uid}>
                <Table.Cell collapsing>
                    <Popup content='Disble/Expire node' trigger={<Checkbox slider value='enabled' checked={this.state.value === 'enabled'} onChange={this.handleChange} />} />
                </Table.Cell>
                <Table.Cell disabled={this.state.value === 'disabled'}>{this.props.uid}</Table.Cell>
                <Table.Cell disabled={this.state.value === 'disabled'}>{this.props.name}</Table.Cell>
                <Table.Cell disabled={this.state.value === 'disabled'}>{this.props.address}</Table.Cell>
            </Table.Row>
        )
    }
}

export class Registry extends Component {

    constructor(props) {
        super(props)
        this.state = { rows: [], open: false }
    }

    componentDidMount() {
        /*
            Fetch all the services from the Restful endpoint
            and create all the rows.
        */
        fetch(this.props.service + "/services")
            .then(data => {
                return data.json()
            })
            .then(json => {
                let rows = []

                json.nodes.forEach(node => {
                    rows.push(<RegistryRow key={node.uid} service={this.props.service} uid={node.uid} name={node.name} address={node.address} />)
                });

                return rows
            })
            .then(rows => {
                this.setState({ rows: rows })
            })
            .catch(err => { console.error(err) })
    }

    render() {
        /*
            Render the table with all the services that we received from the Restful endpoint.
        */
        return <div>
            <Table id="registry-table" celled color="blue" striped textAlign="left" columns="3" definition compact>
                <Table.Header>
                    <Table.Row>
                        <Table.HeaderCell>Disable/Expire</Table.HeaderCell>
                        <Table.HeaderCell>UID</Table.HeaderCell>
                        <Table.HeaderCell>Name</Table.HeaderCell>
                        <Table.HeaderCell>Address</Table.HeaderCell>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    {this.state.rows}
                </Table.Body>
                <Table.Footer fullWidth>
                    <Table.Row>
                        <Table.HeaderCell />
                        <Table.HeaderCell colSpan='4'>
                            <Modal dimmer open={this.state.open} onOpen={() => this.setState({ open: true })} onClose={() => this.setState({ open: false })} trigger={<Button floated='right' icon labelPosition='left' primary size='small'><Icon name='add' />Add service</Button>}>
                                <Modal.Header>Register a new service</Modal.Header>
                                <Modal.Content><AddNodeForm  service={this.props.service} /></Modal.Content>
                                <Modal.Actions></Modal.Actions>
                            </Modal>
                        </Table.HeaderCell>
                    </Table.Row>
                </Table.Footer>
            </Table>
        </div >
    }

}


Registry.defaultProps = {
    service: "http://localhost:8080/api/v1"
}