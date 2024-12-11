import React from "react";
import { connect } from "react-redux";
import { RootState } from "redux/reducers";
import {
  DataTable,
  DataTableCustomBatchActionsData,
  DataTableCustomBatchActionsProps,
  DenormalizedRow,
} from "carbon-components-react";
import { IoBan, IoDownloadOutline, IoTrashOutline } from "react-icons/io5";
import { ReactDivAttr, ShapeOf } from "carbon-components-react/typings/shared";
import TornjakHelper from "components/tornjak-helper";
import { env } from "../env";

const {
  TableToolbar,
  TableToolbarSearch,
  TableToolbarContent,
  TableBatchActions,
  TableBatchAction,
} = DataTable;

const Auth_Server_Uri = env.REACT_APP_AUTH_SERVER_URI;

type TableToolBarProps = {
  deleteEntity?: (selectedRows: readonly DenormalizedRow[]) => string | void;
  banEntity?: (selectedRows: readonly DenormalizedRow[]) => string | void;
  downloadEntity?: (selectedRows: readonly DenormalizedRow[]) => void;
  onInputChange: (event: React.SyntheticEvent<HTMLInputElement, Event>) => void;
  getBatchActionProps: <E extends object = ReactDivAttr>(
    data?: ShapeOf<DataTableCustomBatchActionsData, E>
  ) => ShapeOf<DataTableCustomBatchActionsProps, E>;
  selectedRows: readonly DenormalizedRow[];
  globalUserRoles: string[];
};

class TableToolBar extends React.Component<TableToolBarProps> {
  private tornjakHelper: TornjakHelper;

  constructor(props: TableToolBarProps) {
    super(props);
    this.tornjakHelper = new TornjakHelper(props);
  }

  handleBatchAction = (
    action?: (selectedRows: readonly DenormalizedRow[]) => void
  ) => {
    if (action) {
      action(this.props.selectedRows);
    }
    this.props.getBatchActionProps().onCancel();
  };

  render() {
    const { deleteEntity, banEntity, downloadEntity, onInputChange, getBatchActionProps, globalUserRoles } = this.props;

    return (
      <TableToolbar>
        <TableToolbarContent>
          <TableToolbarSearch onChange={onInputChange} />
        </TableToolbarContent>
        <TableBatchActions {...getBatchActionProps()}>
          {(deleteEntity &&
            (this.tornjakHelper.checkRolesAdminUser(globalUserRoles) || !Auth_Server_Uri)) && (
            <TableBatchAction
              renderIcon={IoTrashOutline}
              iconDescription="Delete"
              onClick={() => this.handleBatchAction(deleteEntity)}
            >
              Delete
            </TableBatchAction>
          )}
          {downloadEntity && (
            <TableBatchAction
              renderIcon={IoDownloadOutline}
              iconDescription="Download"
              onClick={() => this.handleBatchAction(downloadEntity)}
            >
              Export to Json
            </TableBatchAction>
          )}
          {(banEntity &&
            (this.tornjakHelper.checkRolesAdminUser(globalUserRoles) || !Auth_Server_Uri)) && (
            <TableBatchAction
              renderIcon={IoBan}
              iconDescription="Ban"
              onClick={() => this.handleBatchAction(banEntity)}
            >
              Ban
            </TableBatchAction>
          )}
        </TableBatchActions>
      </TableToolbar>
    );
  }
}

const mapStateToProps = (state: RootState) => ({
  globalUserRoles: state.auth.globalUserRoles,
});

export default connect(mapStateToProps)(TableToolBar);
export { TableToolBar };