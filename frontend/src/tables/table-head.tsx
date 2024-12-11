import React from "react";
import {
  DataTable,
  DataTableCustomHeaderData,
  DataTableCustomHeaderProps,
  DataTableCustomSelectionData,
  DataTableCustomSelectionProps,
  DataTableHeader,
  DataTableRow,
} from "carbon-components-react";
import { ReactAttr, ShapeOf } from "carbon-components-react/typings/shared";

const { TableHead, TableRow, TableSelectAll, TableHeader } = DataTable;

// Head takes in:
// - getSelectionProps: Function for selecting all rows from DataTable
// - headers: Header data of the table
// - getHeaderProps: Function to get header properties from DataTable
// Returns the header of the table for the specified entity
type HeadProp = {
  headers: DataTableHeader<string>[];
  getSelectionProps: <E extends object = {}>(
    data?:
      | ShapeOf<DataTableCustomSelectionData<DataTableRow<string>>, E>
      | undefined
  ) => ShapeOf<DataTableCustomSelectionProps<DataTableRow<string>>, E> | ShapeOf<DataTableCustomSelectionProps<never>, E>;
  getHeaderProps: <E extends object = ReactAttr<HTMLElement>>(
    data: ShapeOf<DataTableCustomHeaderData<DataTableHeader<string>>, E>
  ) => ShapeOf<DataTableCustomHeaderProps<DataTableHeader<string>>, E>;
};

class Head extends React.Component<HeadProp> {
  render() {
    const { headers, getSelectionProps, getHeaderProps } = this.props;

    return (
      <TableHead>
        <TableRow>
          <TableSelectAll {...getSelectionProps()} />
          {headers.map((header) => (
            <TableHeader key={header.key} {...getHeaderProps({ header })}>
              {header.header}
            </TableHeader>
          ))}
        </TableRow>
      </TableHead>
    );
  }
}

export default Head;