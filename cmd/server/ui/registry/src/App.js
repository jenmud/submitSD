import 'semantic-ui-css/semantic.min.css'
import './App.css';
import { Container, Header, Icon, Table } from 'semantic-ui-react';
import axios from 'axios';
import $ from 'jquery';
import 'datatables.net-se';

function App() {
  return (
    <div className="App">
      <Container className="app-container">
        <Header><Icon name="registered outline">egistry</Icon></Header>

        <Table celled>
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell>UID</Table.HeaderCell>
              <Table.HeaderCell>Name</Table.HeaderCell>
              <Table.HeaderCell>Address</Table.HeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            <Table.Row>
              <Table.Cell>demo-uid-1234</Table.Cell>
              <Table.Cell>Demo.srv</Table.Cell>
              <Table.Cell>localhost:1234</Table.Cell>
            </Table.Row>
          </Table.Body>
        </Table>

      </Container>
    </div>
  );
}

export default App;
