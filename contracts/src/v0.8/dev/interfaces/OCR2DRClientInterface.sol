// SPDX-License-Identifier: MIT
pragma solidity ^0.8.6;

/**
 * @title OCR2DR client interface.
 */
interface OCR2DRClientInterface {
  /**
   * @notice OCR2DR response handler called by the designated oracle.
   * @param requestId The requestId returned by OCR2DRClient.sendRequest().
   * @param response Aggregated response from the user code.
   * @param err Aggregated error either from the user code or from the execution pipeline.
   * Either response or error parameter will be set, but never both.
   */
  function handleOracleFulfillment(
    bytes32 requestId,
    bytes memory response,
    bytes memory err
  ) external;
}
