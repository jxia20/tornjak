import { Component } from 'react';
import './style.css';
import { RootState } from 'redux/reducers';
import { connect } from 'react-redux';
import { InlineLoading } from 'carbon-components-react';
import TornjakApi from './tornjak-api-helpers';
import {
  spireHealthCheckFunc,
  spireHealthCheckingFunc,
} from 'redux/actions';

type SpireHealthCheckProp = {
  // dispatches a payload for the status of SPIRE health as a boolean has a return type of void
  spireHealthCheckFunc: (globalSpireHealthCheck: boolean) => void,
  // the status of SPIRE health
  globalSpireHealthCheck: boolean,
  // dispatches a payload for the loading state of SPIRE health as a boolean and has a return type of void
  spireHealthCheckingFunc: (globalSpireHealthCheck: boolean) => void,
  // the loading state of SPIRE health
  globalSpireHealthChecking: boolean,
}

type SpireHealthCheckState = {
  // timer to check SPIRE health
  timer: NodeJS.Timeout | Date | null,
}

class SpireHealthCheck extends Component<SpireHealthCheckProp, SpireHealthCheckState> {
  TornjakApi: TornjakApi;
  constructor(props: SpireHealthCheckProp) {
    super(props);
    this.TornjakApi = new TornjakApi(props);
    this.state = {
      timer: null,
    };
  }

  componentDidMount() {
    this.startTimer();
    this.TornjakApi.spireHealthCheck(this.props.spireHealthCheckFunc, this.props.spireHealthCheckingFunc);
  }

  componentDidUpdate(prevProps: SpireHealthCheckProp, prevState: SpireHealthCheckState) {
    if (prevState.timer !== this.state.timer) {
      // timer ended, check status of SPIRE
      this.TornjakApi.spireHealthCheck(this.props.spireHealthCheckFunc, this.props.spireHealthCheckingFunc);
    }
  }

  startTimer = () => {
    const timer = setTimeout(() => {
      this.setState({ timer: new Date() });
      this.startTimer(); // restart timer
    }, 60000); // default time to check SPIRE health in milliseconds
    this.setState({ timer });
  };

  render() {
    const success = this.props.globalSpireHealthCheck, checking = this.props.globalSpireHealthChecking;
    let spireStatus =
      <div>
        <InlineLoading
          status={success ? 'finished' : 'error'}
          description={success ? 'Active' : 'InActive'} />
      </div>
    return (
      <div className="health-check">
        <div className='health-status-check-container'>
          <div className='health-status-check-title'>
            <h6>SPIRE: </h6>
          </div>
          <div className='health-status-check-icon'>
            {!checking &&
              <InlineLoading
                status="active"
                description="Checking" />
            }
            {checking && spireStatus}
          </div>
        </div>
      </div>
    );
  }
}

const mapStateToProps = (state: RootState) => ({
  globalSpireHealthCheck: state.servers.globalSpireHealthCheck,
  globalSpireHealthChecking: state.servers.globalSpireHealthChecking,
})

export default connect(
  mapStateToProps,
  { spireHealthCheckFunc, spireHealthCheckingFunc }
)(SpireHealthCheck)

export { SpireHealthCheck }
