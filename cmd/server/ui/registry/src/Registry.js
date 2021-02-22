import React, { Component } from 'react';
import { Table } from 'semantic-ui-react';

import axios from 'axios';
import $ from 'jquery';
import 'datatables.net-se';
import { Node } from './Node.js';

export class Registry extends Component {

    constructor(props) {
        super(props)
    }

    componentDidMount() {
        var t = $("table#registry-table").DataTable();
        this.services()
            .then(resp => resp.data)
            .then(json => {
                json["nodes"].forEach(node => {
                    t.row.add(
                        [
                            node.uid,
                            node.name,
                            node.address,
                        ]
                    ).draw();
                });
            }).catch(err => console.error(err));
    }

    services() {
        return axios.get(this.props.service + "/services");
    }

    node(id) {
        return axios.get(this.props.service + "/node/" + id)
    }

    render() {
        return <div>
            <Table id="registry-table" celled color="blue" striped textAlign="left" columns="3">
                <Table.Header>
                    <Table.Row>
                        <Table.HeaderCell>UID</Table.HeaderCell>
                        <Table.HeaderCell>Name</Table.HeaderCell>
                        <Table.HeaderCell>Address</Table.HeaderCell>
                    </Table.Row>
                </Table.Header>
            </Table>
        </div >
    }

}


Registry.defaultProps = {
    service: "http://localhost:8080/api/v1"
}