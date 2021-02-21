import 'fomantic-ui-css/semantic.css';
import { Container, Header, Table } from 'semantic-ui-react';

function App() {
  return (
    <div className="App">
      <Container>
        <Header>Registry</Header>

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
