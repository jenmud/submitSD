import React, { Component } from 'react';
import { Table } from 'semantic-ui-react';

import axios from 'axios';

export class Registry extends Component {

    constructor(props) {
        super(props)
    }

    rows() {
        /*
            Fetch all the services from the Restful endpoint
            and create all the rows.
        */
        let resp = async () => await fetch(this.props.service + "/services")
        if (resp.ok) {
            return resp.json()
        }

        return []
    }

    render() {
        const rows = this.rows()

        //var resp = async () => await axios.get(this.props.service + "/services")
        //    .then(resp => resp.data)
        //    .then(json => {
        //        json["nodes"].forEach(node => {
        //            console.debug(node)
        //            rows.push(
        //                <Table.Row>
        //                    <Table.Cell>{node.uid}</Table.Cell>
        //                    <Table.Cell>{node.name}</Table.Cell>
        //                    <Table.Cell>{node.address}</Table.Cell>
        //                </Table.Row>
        //            )
        //        })
        //    })
        //    .catch(err => console.error(err));

        console.log(rows)
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
                    {rows}
                </Table.Body>
            </Table>
        </div >
    }

}


Registry.defaultProps = {
    service: "http://localhost:8080/api/v1"
}