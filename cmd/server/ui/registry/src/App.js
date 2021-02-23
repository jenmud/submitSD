import 'semantic-ui-css/semantic.min.css'
import './App.css';
import { Container, Header, Icon, Tab, Placeholder } from 'semantic-ui-react';

import { Registry } from './Registry.js';

function RegistryIcon() {
  return (
    <div className="app-icon">
      <Header><Icon name="registered outline">egistry</Icon></Header>
    </div>
  )
}

const panes = [
  { menuItem: 'Services', render: () => <Tab.Pane>
        <Registry />
    </Tab.Pane>
  },
  { menuItem: 'About', render: () => <Tab.Pane>
        <Placeholder>
          <Placeholder.Line />
          <Placeholder.Line />
          <Placeholder.Line />
          <Placeholder.Line />
          <Placeholder.Line />
        </Placeholder>
    </Tab.Pane>
  },
]

function App() {
  return (
    <div className="App">
      <Container className="app-container">
      <RegistryIcon />
        <Tab panes={panes} />
      </Container>
    </div>
  );
}

export default App;
