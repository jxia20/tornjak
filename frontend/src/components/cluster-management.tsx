import { Component } from 'react';
import { Tabs, TabList, Tab, TabPanels, TabPanel } from '@carbon/react';
import ClusterCreate from './cluster-create';
import ClusterEdit from './cluster-edit';
import { connect } from 'react-redux';
import IsManager from './is_manager';
import TornjakApi from './tornjak-api-helpers';
import './style.css';
import {
  clusterTypeInfoFunc,
  serverSelectedFunc,
  selectorInfoFunc,
  agentsListUpdateFunc,
  tornjakMessageFunc,
  tornjakServerInfoUpdateFunc,
  serverInfoUpdateFunc
} from 'redux/actions';
import { RootState } from 'redux/reducers';
import {
  AgentLabels,
  AgentsList,
  ServerInfo,
  TornjakServerInfo,
  DebugServerInfo
} from './types';
import { toast } from 'react-toastify';
// import PropTypes from "prop-types"; // needed for testing will be removed on last pr

type ClusterManagementProp = {
  globalDebugServerInfo: DebugServerInfo,
  agentsListUpdateFunc: (globalAgentsList: AgentsList[]) => void,
  tornjakMessageFunc: (globalErrorMessage: string) => void,
  tornjakServerInfoUpdateFunc: (globalTornjakServerInfo: TornjakServerInfo) => void,
  serverInfoUpdateFunc: (globalServerInfo: ServerInfo) => void,
  globalServerSelected: string,
  globalErrorMessage: string,
  globalTornjakServerInfo: TornjakServerInfo,
  globalServerInfo: ServerInfo,
  globalClusterTypeInfo: string[],
  globalAgentsList: AgentsList[],
}

type ClusterManagementState = {
  clusterTypeList: string[],
  agentsList: AgentLabels[],
  agentsListDisplay: string,
  clusterTypeManualEntryOption: string,
  selectedServer: string,
  isLoading: boolean, // New state for managing loading indicator
}

class ClusterManagement extends Component<ClusterManagementProp, ClusterManagementState> {
  TornjakApi: TornjakApi;
  constructor(props: ClusterManagementProp) {
    super(props);
    this.TornjakApi = new TornjakApi(props);
    this.prepareClusterTypeList = this.prepareClusterTypeList.bind(this);
    this.prepareAgentsList = this.prepareAgentsList.bind(this);
    this.handleTabSelect = this.handleTabSelect.bind(this); // Added binding for handleTabSelect
    this.state = {
      clusterTypeList: [],
      agentsList: [],
      agentsListDisplay: "Select Agents",
      clusterTypeManualEntryOption: "----Select this option and Enter Custom Cluster Type Below----",
      selectedServer: "",
      isLoading: true, // Initialize loading state
    };
  }

  componentDidMount() {
    if (IsManager) {
      if (this.props.globalServerSelected !== "" && (this.props.globalErrorMessage === "OK" || this.props.globalErrorMessage === "")) {
        this.TornjakApi.populateAgentsUpdate(this.props.globalServerSelected, this.props.agentsListUpdateFunc, this.props.tornjakMessageFunc);
        this.TornjakApi.populateTornjakServerInfo(this.props.globalServerSelected, this.props.tornjakServerInfoUpdateFunc, this.props.tornjakMessageFunc);
        this.setState({ selectedServer: this.props.globalServerSelected });
        this.prepareClusterTypeList();
        this.prepareAgentsList();
      }
    } else {
      this.TornjakApi.populateLocalAgentsUpdate(this.props.agentsListUpdateFunc, this.props.tornjakMessageFunc);
      this.TornjakApi.populateLocalTornjakServerInfo(this.props.tornjakServerInfoUpdateFunc, this.props.tornjakMessageFunc);
      this.TornjakApi.populateServerInfo(this.props.globalTornjakServerInfo, this.props.serverInfoUpdateFunc);
      this.prepareClusterTypeList();
      this.prepareAgentsList();
    }
    this.setState({ isLoading: false }); // Set loading to false after data fetch
  }

  componentDidUpdate(prevProps: ClusterManagementProp, prevState: ClusterManagementState) {
    if (IsManager) {
      if (prevProps.globalServerSelected !== this.props.globalServerSelected) {
        this.setState({ selectedServer: this.props.globalServerSelected });
      }
      if (prevProps.globalDebugServerInfo !== this.props.globalDebugServerInfo) {
        this.prepareAgentsList();
      }
    } else {
      if (prevProps.globalDebugServerInfo !== this.props.globalDebugServerInfo) {
        this.prepareAgentsList();
      }
    }
  }

  prepareClusterTypeList(): void {
    let localClusterTypeList = [this.state.clusterTypeManualEntryOption];
    for (let i = 0; i < this.props.globalClusterTypeInfo.length; i++) {
      localClusterTypeList.push(this.props.globalClusterTypeInfo[i]);
    }
    this.setState({ clusterTypeList: localClusterTypeList });
  }

  prepareAgentsList(): void {
    const prefix = "spiffe://";
    let localAgentsIdList: AgentLabels[] = [];
    if (this.props.globalAgentsList === undefined) {
      return;
    }
    for (let i = 0; i < this.props.globalAgentsList.length; i++) {
      localAgentsIdList[i] = { label: "" };
      localAgentsIdList[i]["label"] = prefix + this.props.globalAgentsList[i].id.trust_domain + this.props.globalAgentsList[i].id.path;
    }
    this.setState({
      agentsList: localAgentsIdList,
    });
  }

  handleTabSelect(): void {
    toast.dismiss();
    this.setState({ agentsListDisplay: "Select Agents" }); // Reset agents list display on tab select
  }

  render() {
    if (this.state.isLoading) {
      return <div>Loading...</div>; // Loading indicator
    }
    return (
      <div className="cluster-management-tabs" data-test="cluster-management">
        <Tabs>
          <TabList aria-label="hi">
            <Tab
              className="cluster-management-tab1"
              id="tab-1"
              onClick={this.handleTabSelect}
            >
              Create Cluster
            </Tab>
            <Tab
              id="tab-2"
              onClick={this.handleTabSelect}
            >
              Edit Cluster
            </Tab>
          </TabList>
          <TabPanels>
            <TabPanel>
              <ClusterCreate
                clusterTypeList={this.state.clusterTypeList}
                agentsList={this.state.agentsList}
              />
            </TabPanel>
            <TabPanel>
              <ClusterEdit
                clusterTypeList={this.state.clusterTypeList}
                agentsList={this.state.agentsList}
              />
            </TabPanel>
          </TabPanels>
        </Tabs>
      </div>
    );
  }
}

const mapStateToProps = (state: RootState) => ({
  globalClusterTypeInfo: state.clusters.globalClusterTypeInfo,
  globalServerSelected: state.servers.globalServerSelected,
  globalAgentsList: state.agents.globalAgentsList,
  globalServerInfo: state.servers.globalServerInfo,
  globalTornjakServerInfo: state.servers.globalTornjakServerInfo,
  globalErrorMessage: state.tornjak.globalErrorMessage,
  globalDebugServerInfo: state.servers.globalDebugServerInfo,
});

export default connect(
  mapStateToProps,
  { clusterTypeInfoFunc, serverSelectedFunc, selectorInfoFunc, agentsListUpdateFunc, tornjakMessageFunc, tornjakServerInfoUpdateFunc, serverInfoUpdateFunc }
)(ClusterManagement);

export { ClusterManagement };
