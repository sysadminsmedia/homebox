# External Label Service

You can use an external web service to generate asset and location labels in homebox. This is useful if you have custom requirements for your labels and are happy to spin up a web service that can accept incoming requests and return an image file for homebox to use.

::: info "Note" 

This service is not called to generate sheets of labels accessed via the label generator function. It is used when creating labels from an item or location.

:::

## Configuration

The extenal service is configured using the `HBOX_LABEL_MAKER_LABEL_SERVICE_URL` enviroment variable.

## Request

The service is called using an **HTTP `GET` request**.  All parameters are passed as part of the **query string**.

#### Headers

- **User-Agent**: Homebox-LabelMaker/1.0

- **Accept**: image/*

#### Parameters

| Parameter             | Type   | Description                                  | Value                                                                 |
| --------------------- | ------ | -------------------------------------------- | --------------------------------------------------------------------- |
| AdditionalInformation | string | Extra free text to include on the label.     | `HBOX_LABEL_MAKER_ADDITIONAL_INFORMATION`                             |
| ComponentPadding      | int    | Padding around label components (pixels).    | `HBOX_LABEL_MAKER_PADDING`                                            |
| DescriptionFontSize   | float  | Font size for the description text.          |                                                                       |
| DescriptionText       | string | Descriptive text, can be multi-line.         | Item name or "Homebox Location"                                       |
| Dpi                   | float  | Rendering resolution (dots per inch).        |                                                                       |
| DynamicLength         | bool   | Whether the label length should auto-adjust. | `HBOX_LABEL_MAKER_DYNAMIC_LENGTH`                                     |
| Height                | int    | Label height in pixels.                      | `HBOX_LABEL_MAKER_HEIGHT`                                             |
| Margin                | int    | Margin around the label in pixels.           | `HBOX_LABEL_MAKER_MARGIN`                                             |
| QrSize                | int    | Size of the QR code element in pixels.       |                                                                       |
| TitleFontSize         | float  | Font size for the title text.                |                                                                       |
| TitleText             | string | Main label title (e.g. product code).        | Asset ID or Location Name                                             |
| URL                   | string | URL to be encoded into the QR code.          | Generated based on the configured homebox URL and Asset / Location ID |
| Width                 | int    | Label width in pixels.                       | `HBOX_LABEL_MAKER_WIDTH`                                              |

## Response

The external service should respond with the following specifications;

- **Size:** Less than or equal to `HBOX_WEB_MAX_UPLOAD_SIZE` (Default: 10Mb)

- **Content-Type**: Specified in the response header should be of the type image/*

- **Time**: Within the time specified in `HBOX_LABEL_MAKER_LABEL_SERVICE_TIMEOUT` (Default 30s)


