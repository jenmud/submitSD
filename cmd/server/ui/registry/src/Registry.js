import React, { Component } from 'react';
import { Table } from 'semantic-ui-react';

export class Registry extends Component {

    constructor(props) {
        super(props)
        this.state = { rows: [] }
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
                    rows.push(
                        <Table.Row id={node.uid} key={node.uid}>
                            <Table.Cell>{node.uid}</Table.Cell>
                            <Table.Cell>{node.name}</Table.Cell>
                            <Table.Cell>{node.address}</Table.Cell>
                        </Table.Row>
                    )
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
            <Table id="registry-table" celled color="blue" striped textAlign="left" columns="3">
                <Table.Header>
                    <Table.Row>
                        <Table.HeaderCell>UID</Table.HeaderCell>
                        <Table.HeaderCell>Name</Table.HeaderCell>
                        <Table.HeaderCell>Address</Table.HeaderCell>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    {this.state.rows}
                </Table.Body>
            </Table>
        </div >
    }

}


Registry.defaultProps = {
    service: "http://localhost:8080/api/v1"
}