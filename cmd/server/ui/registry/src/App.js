import 'semantic-ui-css/semantic.min.css'
import './App.css';
import { Container, Header, Icon, Table } from 'semantic-ui-react';

import { Registry } from './Registry.js';

function RegistryIcon() {
  return (
    <div className="app-icon">
      <Header><Icon name="registered outline">egistry</Icon></Header>
    </div>
  )
}

function App() {
  return (
    <div className="App">
      <Container className="app-container">
        <RegistryIcon />
        <Registry />
      </Container>
    </div>
  );
}

export default App;
